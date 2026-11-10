package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"example.com/go-service-kit/db/redis"
	goredis "github.com/redis/go-redis/v9"
)

var Prefix string

const (
	KeyAdminLastToken      = "%s:admin:token:last:%d"
	KeyAdminLInvalidToken  = "%s:admin:token:invalid:%s"
	KeyAdminTokenRevokedAt = "%s:admin:token:revoked_at:%d"
)

// SetAdminLastJwtToken 记录当前有效令牌，写入失败不能返回登录成功。
func SetAdminLastJwtToken(ctx context.Context, userID int64, jti string, ttl time.Duration) error {
	return redis.RDB.Set(ctx, fmt.Sprintf(KeyAdminLastToken, Prefix, userID), jti, ttl).Err()
}

func GetAdminLastJwtToken(ctx context.Context, userID int64) (string, error) {
	value, err := redis.RDB.Get(ctx, fmt.Sprintf(KeyAdminLastToken, Prefix, userID)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", nil
	}
	return value, err
}

// AddInvalidAdminToken 将撤销状态保存在Redis，跨重启及实例生效。
func AddInvalidAdminToken(ctx context.Context, jti string, ttl time.Duration) error {
	return redis.RDB.Set(ctx, fmt.Sprintf(KeyAdminLInvalidToken, Prefix, jti), "1", ttl).Err()
}

func ExistInvalidAdminToken(ctx context.Context, jti string) (bool, error) {
	count, err := redis.RDB.Exists(ctx, fmt.Sprintf(KeyAdminLInvalidToken, Prefix, jti)).Result()
	return count > 0, err
}

const revokeScript = `
local previous = redis.call('GET', KEYS[1])
if not previous or tonumber(previous) < tonumber(ARGV[1]) then
  redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
end
return 1
`

// SetAdminTokenRevokedAt 不允许较早的并发撤销覆盖较新的撤销时间。
func SetAdminTokenRevokedAt(ctx context.Context, adminID int64, revokedAtUnix int64, ttl time.Duration) error {
	key := fmt.Sprintf(KeyAdminTokenRevokedAt, Prefix, adminID)
	return redis.RDB.Eval(ctx, revokeScript, []string{key}, revokedAtUnix, ttl.Milliseconds()).Err()
}

func GetAdminTokenRevokedAt(ctx context.Context, adminID int64) (int64, error) {
	value, err := redis.RDB.Get(ctx, fmt.Sprintf(KeyAdminTokenRevokedAt, Prefix, adminID)).Result()
	if errors.Is(err, goredis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(value, 10, 64)
}

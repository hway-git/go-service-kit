package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"example.com/go-service-kit/db/redis"
	goredis "github.com/redis/go-redis/v9"
)

type EmailLoginChallenge struct {
	AdminID  int    `json:"admin_id"`
	Version  string `json:"version"`
	Digest   string `json:"digest"`
	Attempts int    `json:"attempts"`
}

const consumeEmailChallenge = `
local raw = redis.call('GET', KEYS[1])
if not raw then
  return false
end
local challenge = cjson.decode(raw)
if challenge.digest == ARGV[1] then
  redis.call('DEL', KEYS[1])
  return raw
end
challenge.attempts = challenge.attempts + 1
local ttl = redis.call('PTTL', KEYS[1])
if challenge.attempts >= tonumber(ARGV[2]) or ttl <= 0 then
  redis.call('DEL', KEYS[1])
else
  redis.call('PSETEX', KEYS[1], ttl, cjson.encode(challenge))
end
return false
`

func ReserveEmailLoginSend(ctx context.Context, adminID int, cooldown time.Duration) (bool, error) {
	key := fmt.Sprintf("%s:admin:login:email-cooldown:%d", Prefix, adminID)
	return redis.RDB.SetNX(ctx, key, "1", cooldown).Result()
}

func emailChallengeKey(id string) string {
	return Prefix + ":admin:login:email:" + id
}

func SaveEmailLoginChallenge(ctx context.Context, id string, challenge EmailLoginChallenge, ttl time.Duration) error {
	data, err := json.Marshal(challenge)
	if err != nil {
		return err
	}
	return redis.RDB.Set(ctx, emailChallengeKey(id), data, ttl).Err()
}

func DeleteEmailLoginChallenge(ctx context.Context, id string) error {
	return redis.RDB.Del(ctx, emailChallengeKey(id)).Err()
}

// ConsumeEmailLoginChallenge 原子地核销验证码或计入一次错误尝试。
func ConsumeEmailLoginChallenge(ctx context.Context, id, digest string, maxAttempts int) (*EmailLoginChallenge, error) {
	data, err := redis.RDB.Eval(ctx, consumeEmailChallenge, []string{emailChallengeKey(id)}, digest, maxAttempts).Text()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var challenge EmailLoginChallenge
	if err := json.Unmarshal([]byte(data), &challenge); err != nil {
		return nil, err
	}
	return &challenge, nil
}

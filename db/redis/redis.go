package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"example.com/go-service-kit/utils/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	RDB redis.UniversalClient
	ctx context.Context = context.TODO()
)

func Init(addr, password string, poolSize int, db int, tlsEnabled bool) {
	// 连接串
	options := &redis.UniversalOptions{
		Addrs:    []string{addr}, // 单点地址
		Password: password,       // 密码
		DB:       db,             // DB（cluster 模式会被忽略）
		PoolSize: poolSize,       // 连接池
	}
	// 是否启用TLS连接
	if tlsEnabled {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	RDB = redis.NewUniversalClient(options)
	if _, err := RDB.Ping(ctx).Result(); err != nil {
		log.Fatalf("Redis connect failed (single): %v", err)
	} else {
		logger.Info("Redis连接初始化完成，使用单点模式")
	}
}

// InitCluster 初始化 Redis Cluster
func InitCluster(addrs []string, password string, poolSize int, tlsEnabled bool) {
	options := &redis.ClusterOptions{
		Addrs:    addrs,
		Password: password,
		PoolSize: poolSize,
	}
	if tlsEnabled {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	RDB = redis.NewClusterClient(options)
	if _, err := RDB.Ping(ctx).Result(); err != nil {
		log.Fatalf("Redis connect failed (cluster): %v", err)
	} else {
		logger.Info("Redis连接初始化完成，使用集群模式", zap.Int("node_count", len(addrs)))
	}
}

func Close() {
	RDB.Close()
}

// GetClientID 获取当前连接的客户端 ID，用于验证连接是否变化
func GetClientID() (int64, error) {
	return RDB.ClientID(ctx).Result()
}

// GetPoolStats 获取连接池统计信息
// stats := redis.GetPoolStats()
// fmt.Printf("总连接数: %d, 空闲: %d, 使用中: %d\n", stats.TotalConns, stats.IdleConns, stats.StaleConns)
func GetPoolStats() *redis.PoolStats {
	return RDB.PoolStats()
}

func Test() {
	fmt.Println(HMSet("hmset", map[string]string{"33": "44", "55": "66"}))
	fmt.Println(HMGet("hmset", []string{"33", "44", "55"}...))
}

func GetInt(key string) (int, error) {
	return RDB.Get(ctx, key).Int()
}

func GetInt64(key string) (int64, error) {
	return RDB.Get(ctx, key).Int64()
}

func GetString(key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

func GetByte(key string) ([]byte, error) {
	return RDB.Get(ctx, key).Bytes()
}

func Exists(key string) (int64, error) {
	return RDB.Exists(ctx, key).Result()
}

func Expire(key string, expire time.Duration) (bool, error) {
	return RDB.Expire(ctx, key, time.Duration(expire)).Result()
}

// 使用变长参数来实现默认值
func Set(key string, value interface{}, expire ...time.Duration) error {
	exp := time.Duration(0) // 默认值为 0，即无过期时间
	if len(expire) > 0 {
		exp = expire[0] // 如果传递了值，则使用该值
	}
	return RDB.Set(ctx, key, value, exp).Err()
}

// 存储字节数组
func SetByte(key string, value []byte, expire ...time.Duration) error {
	exp := time.Duration(0) // 默认值为 0，即无过期时间
	if len(expire) > 0 {
		exp = expire[0] // 如果传递了值，则使用该值
	}
	return RDB.Set(ctx, key, value, exp).Err()
}

// =======================    Hash  ============
// HSet，支持单个或多个 field-value 对
// 示例: 单个: HSet("key", "field1", "value1")
// 示例: 批量：HSet("key", "field1", "value1", "field2", "value2")
func HSet(key string, args ...interface{}) (int64, error) {
	return RDB.HSet(ctx, key, args...).Result()
}

func HGet(key string, field string) (string, error) {
	return RDB.HGet(ctx, key, field).Result()
}

func HDel(key string, fields ...string) (int64, error) {
	return RDB.HDel(ctx, key, fields...).Result()
}

// HMSet 是 HSet 的别名，但官方推荐用 HSet 代替 HMSet
func HMSet(key string, values ...interface{}) (bool, error) {
	return RDB.HMSet(ctx, key, values...).Result()
}

func HMGet(key string, fields ...string) []interface{} {
	return RDB.HMGet(ctx, key, fields...).Val()
}

func HGetAll(key string) map[string]string {
	return RDB.HGetAll(ctx, key).Val()
}

// HINCRBY 专门用于对 Hash 字段进行原子增加
func HIncrByInt64(key, field string, increment int64) (int64, error) {
	return RDB.HIncrBy(ctx, key, field, increment).Result() // 原子增加整数字段
}

// 浮点数原子增加
func HIncrByFloat(key, field string, increment float64) (float64, error) {
	return RDB.HIncrByFloat(ctx, key, field, increment).Result() // 原子增加浮点数字段
}

func HGetByBytes(key string, field string) (reply []byte, err error) {
	return RDB.HGet(ctx, key, field).Bytes()
}

// 使用管道获取hash多个字段的值
func PipeHGet(key string, fields []string) (map[string]string, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("fields is empty")
	}
	pipe := RDB.Pipeline()                       // 创建管道
	results := make(map[string]*redis.StringCmd) // 用于存储每个键的 Get 命令结果
	for _, field := range fields {               // 将每个键的 Get 命令添加到管道
		results[field] = pipe.HGet(ctx, key, field)
	}
	_, err := pipe.Exec(ctx)            // 执行管道中的所有命令
	if err != nil && err != redis.Nil { // redis.Nil 表示键不存在，不算错误
		return nil, fmt.Errorf("pipeline exec error: %w", err)
	}
	// 获取每个键的值
	values := make(map[string]string)
	for field, cmd := range results {
		val, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				values[field] = "" // 键不存在时返回空字符串
				continue
			}
			return nil, fmt.Errorf("error hget key %s, %s, %w", key, field, err)
		}
		values[field] = val
	}
	return values, nil
}

// PipeHGetField 对多个 Redis Hash Key 批量 HGET 同一 field，减少往返次数（用于商户 finish 状态批量查询）。
func PipeHGetField(keys []string, field string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	pipe := RDB.Pipeline()
	results := make(map[string]*redis.StringCmd, len(keys))
	for _, key := range keys {
		results[key] = pipe.HGet(ctx, key, field)
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("pipeline exec error: %w", err)
	}
	values := make(map[string]string, len(keys))
	for key, cmd := range results {
		val, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				values[key] = ""
				continue
			}
			return nil, fmt.Errorf("error hget key %s, %s, %w", key, field, err)
		}
		values[key] = val
	}
	return values, nil
}

func Delete(key ...string) error {
	return RDB.Del(ctx, key...).Err()
}

func LPush(key string, value string) error {
	return RDB.LPush(ctx, key, value).Err()
}

func RPush(key string, value string) error {
	return RDB.RPush(ctx, key, value).Err()
}

func LPop(key string) (string, error) {
	return RDB.LPop(ctx, key).Result()
}

func RPop(key string) (string, error) {
	return RDB.RPop(ctx, key).Result()
}

func LTrim(key string, start int64, end int64) error {
	return RDB.LTrim(ctx, key, start, end).Err()
}

func LRange(key string, start int64, end int64) ([]string, error) {
	return RDB.LRange(ctx, key, start, end).Result()
}

func LLen(key string) (int64, error) {
	return RDB.LLen(ctx, key).Result()
}

func LRem(key string, value string) (int64, error) {
	return RDB.LRem(ctx, key, 1, value).Result()
}

func Incr(key string) (int64, error) {
	return RDB.Incr(ctx, key).Result()
}

func IncrBy(key string, increment int64) (int64, error) {
	return RDB.IncrBy(ctx, key, increment).Result()
}

func Setnx(key string, val string) (bool, error) {
	return RDB.SetNX(ctx, key, val, 0).Result()
}

func Mget(key ...string) (reply []interface{}, err error) {
	return RDB.MGet(ctx, key...).Result()
}

// 批量设置值  可惜没有办法设置 ttl
func Mset(kv ...interface{}) (reply string, err error) {
	return RDB.MSet(ctx, kv).Result()
}

func Keys(key string) (reply []string, err error) {
	return RDB.Keys(ctx, key).Result()
}

// func ZADD(key string, field string, value int) (reply interface{}, err error) {
// 	return con.Do("zadd", key, value, field)
// }

// func Zrem(key string, field string) (reply interface{}, err error) {
// 	return con.Do("Zrem", key, field)
// }

// func ZREVRANGE(key string, start int, stop int) (reply map[string]string, err error) {
// 	return redis.StringMap(con.Do("ZREVRANGE", key, start, stop, "WITHSCORES"))
// }

// // 删除有序集合末尾几名
// func Zremrangebyrank(key string, start int, stop int) (reply int, err error) {
// 	return redis.Int(con.Do("Zremrangebyrank", key, start, stop))
// }

// // edis Zrangebyscore 返回有序集合中指定分数区间的成员列表。有序集成员按分数值递增(从小到大)次序排列。
// func Zrangebyscore(key string, min string, max string) (reply []string, err error) {
// 	return redis.Strings(con.Do("Zrangebyscore", key, min, max))
// }

// func ZCARD(key string) (reply int, err error) {
// 	return redis.Int(con.Do("ZCARD", key))
// }

// 获取键的过期时间
func TTL(key string) (time.Duration, error) {
	return RDB.TTL(ctx, key).Result()
}

// ###################    管道    ###################

type RedisKv struct {
	Key    string
	Value  string
	Expire time.Duration
}

// 管道批量执行命令
func PipelineSet(keyValues []RedisKv) error {
	if len(keyValues) == 0 {
		return nil
	}
	pipe := RDB.Pipeline()
	for _, item := range keyValues {
		pipe.Set(context.Background(), item.Key, item.Value, item.Expire)
	}
	_, err := pipe.Exec(context.Background())
	return err
}

// 管道批量执行命令
func PipelineSetKeepTTL(keyValues []RedisKv) error {
	if len(keyValues) == 0 {
		return nil
	}
	pipe := RDB.Pipeline()
	for _, item := range keyValues {
		pipe.Set(context.Background(), item.Key, item.Value, RDB.TTL(ctx, item.Key).Val())
	}
	_, err := pipe.Exec(context.Background())
	return err
}

// 管道获取数据
func PipelineGet(keys ...string) ([][]byte, error) {
	pipe := RDB.Pipeline() // 使用管道
	for _, key := range keys {
		pipe.Get(ctx, key).Bytes()
	}
	if _, err := pipe.Exec(ctx); err != nil {
		fmt.Println("Pipeline exec error:", err) // 如果有错误，处理错误
		return nil, err
	}
	// 获取结果
	results := make([][]byte, len(keys))
	var err error
	for i, key := range keys {
		if results[i], err = pipe.Get(ctx, key).Bytes(); err != nil {
			fmt.Println("Pipeline get error:", err) // 如果有错误，处理错误
			return nil, err
		}
	}
	return results, nil
}

// 使用管道获取多个键的值
func GetKeys(keys ...string) (map[string]string, error) {
	pipe := RDB.Pipeline()                       // 创建管道
	results := make(map[string]*redis.StringCmd) // 用于存储每个键的 Get 命令结果
	for _, key := range keys {                   // 将每个键的 Get 命令添加到管道
		results[key] = pipe.Get(ctx, key)
	}
	_, err := pipe.Exec(ctx)            // 执行管道中的所有命令
	if err != nil && err != redis.Nil { // redis.Nil 表示键不存在，不算错误
		return nil, fmt.Errorf("pipeline exec error: %w", err)
	}
	// 获取每个键的值
	values := make(map[string]string)
	for key, cmd := range results {
		val, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				values[key] = "" // 键不存在时返回空字符串
				continue
			}
			return nil, fmt.Errorf("error getting key %s: %w", key, err)
		}
		values[key] = val
	}
	return values, nil
}

// ###################   集合   ###################

// 查看集合元素数量
func SCard(key string) (reply int64, err error) {
	return RDB.SCard(ctx, key).Result()
}

// 集合添加元素
func SAdd(key string, members ...interface{}) (int64, error) {
	return RDB.SAdd(ctx, key, members...).Result()
}

// 随机获取集合1个元素
func SRandMember(key string) (reply string, err error) {
	return RDB.SRandMember(ctx, key).Result()
}

// 随机获取集合指定数量元素
func SRandMemberN(key string, count int64) (reply []string, err error) {
	return RDB.SRandMemberN(ctx, key, count).Result()
}

// 获取集合所有元素
func SMembers(key string) (reply []string, err error) {
	return RDB.SMembers(ctx, key).Result()
}

// 执行 SSCAN 扫描集合
func SScan(key string, cursor uint64, count int64) (reply []string, nextCursor uint64, err error) {
	// keys, nextCursor, err := rdb.SScan(ctx, "myset", cursor, "*", count).Result()
	return RDB.SScan(ctx, key, cursor, "*", count).Result()
}

// 集合移除元素
func SRem(key string, members ...interface{}) (int64, error) {
	return RDB.SRem(ctx, key, members...).Result()
}

// 执行 SSCAN 扫描集合
func Scan(cursor uint64, match string, count int64) (reply []string, nextCursor uint64, err error) {
	return RDB.Scan(ctx, cursor, match, count).Result()
}

// ***********  发布订阅   ***********
// 发布消息
func Publish(channel string, message interface{}) (int64, error) {
	return RDB.Publish(ctx, channel, message).Result()
}

// 订阅频道
func Subscribe(channels ...string) *redis.PubSub {
	return RDB.Subscribe(ctx, channels...)
}

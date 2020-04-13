package clients

import (
	"github.com/go-redis/redis"
)

// NewRedisClient returns a RedisClient.
func NewRedisClient(addr string, passwd string, poolSize int) *redis.Client {
	cli := redis.NewClient(&redis.Options{
		Addr:       addr,
		PoolSize:   poolSize,
		MaxRetries: 2,
		Password:   passwd,
		DB:         0,
	})
	_, err := cli.Ping().Result()
	if err != nil {
		panic(err)
	}
	return cli
}

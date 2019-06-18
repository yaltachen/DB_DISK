package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// var (
// 	client *redis.Client
// )

// func init() {
// 	client = newRedisClient()
// }

// func newRedisClient() *redis.Client {
// 	return redis.NewClient(&redis.Options{
// 		Addr:         "192.168.94.30:6379",
// 		Password:     "foobared",
// 		DB:           0,
// 		DialTimeout:  5000 * time.Millisecond,
// 		WriteTimeout: 5000 * time.Millisecond,
// 		ReadTimeout:  5000 * time.Millisecond,
// 		MinIdleConns: 5,
// 	})
// }

// func RedisPool() *redis.Client {
// 	return client
// }

var (
	pool      *redis.Pool
	redisHost = "192.168.94.30:6379"
	redisPwd  = "foobared"
)

func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   30,
		IdleTimeout: 5000 * time.Millisecond,
		Dial: func() (redis.Conn, error) {
			// 1. 开启连接
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Println(err.Error())
				return nil, err
			}

			// 2. 访问认证
			if _, err = c.Do("AUTH", redisPwd); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("ping")
			return err
		},
	}
}

func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}

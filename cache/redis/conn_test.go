package redis

import (
	"testing"

	"github.com/go-redis/redis"
)

func TestConn(t *testing.T) {
	t.SkipNow()
	var (
		reply interface{}
		err   error
	)
	cmd := RedisPool().Do("PING")
	if reply, err = cmd.Val(), cmd.Err(); err != nil {
		t.Error(err.Error())
	}
	t.Log(reply)
}
func TestRedisConn(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "192.168.94.30:6379",
		Password: "foobared",
		DB:       0, // use default DB
	})

	pong, err := client.Ping().Result()
	t.Log(pong, err)
	// Output: PONG <nil>
}

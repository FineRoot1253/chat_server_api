package lib

import (
	"github.com/gomodule/redigo/redis"
)

var redisPool = newPool()

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "redis:25000")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}

}

// func RedisConnect() {
// 	conn, err := redis.Dial("tcp", ":6379")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer conn.Close()

// 	RedisConn = conn
// 	log.Print("Redis Connection ready")
// }

package redigopool

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	// Pool of redis connections
	Pool *redis.Pool
)

// InitPool replaces init() package method.
// Call InitPool once in client call before accessing Pool
func InitPool(redisHost string, opts ...redis.DialOption) {
	if redisHost == "" {
		redisHost = ":6379"
	}
	Pool = newPool(redisHost, opts)
	cleanupHook()
}

func newPool(server string, opts []redis.DialOption) *redis.Pool {

	return &redis.Pool{

		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server, opts...)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func cleanupHook() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		Pool.Close()
		os.Exit(0)
	}()
}

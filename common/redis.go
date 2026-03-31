package common

import (
	"context"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

// [RDB] 中文名：Redis 客户端实例
// 设计目的：全局变量，保存了与 Redis 服务器的连接池。
// 整个项目（如验证 Token、存取缓存）都通过这个对象与 Redis 通讯。
var RDB *redis.Client

// [RedisEnabled] 中文名：Redis 启用状态开关
// 设计目的：实现“软依赖”设计。如果环境里没配置 Redis，系统依然能靠 SQLite 运行，
// 只是性能稍低。这让项目在不同配置的服务器上都能一键启动。
var RedisEnabled = true

// InitRedisClient This function is called after init()
func InitRedisClient() (err error) {
	if os.Getenv("REDIS_CONN_STRING") == "" {
		RedisEnabled = false
		SysLog("REDIS_CONN_STRING not set, Redis is not enabled")
		return nil
	}
	opt, err := redis.ParseURL(os.Getenv("REDIS_CONN_STRING"))
	if err != nil {
		panic(err)
	}
	RDB = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RDB.Ping(ctx).Result()
	return err
}

func ParseRedisOption() *redis.Options {
	opt, err := redis.ParseURL(os.Getenv("REDIS_CONN_STRING"))
	if err != nil {
		panic(err)
	}
	return opt
}

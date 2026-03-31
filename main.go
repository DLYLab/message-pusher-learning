package main

import (
	"embed"
	"log"
	"message-pusher/channel"
	"message-pusher/common"
	"message-pusher/model"
	"message-pusher/router"
	"os"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

//go:embed web/build
var buildFS embed.FS

//go:embed web/build/index.html
var indexPage []byte

func main() {
	common.SetupGinLog()  // 初始化log文件位置，写入逻辑
	common.SysLog("Message Pusher " + common.Version + " started")

	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize SQL Database
	err := model.InitDB()  // 初始化数据库，有添加mysql，没有则为默认的sqlite
	if err != nil {
		common.FatalLog(err)
	}

	go channel.LoadAsyncMessages()  // 服务器重启，加载未处理的异步消息
	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.FatalLog(err)
		}
	}()

	// Initialize Redis
	err = common.InitRedisClient()  // Redis缓存，当前没有，前端请求直接访问数据库
	if err != nil {
		common.FatalLog(err)
	}

	// Initialize options
	model.InitOptionMap()  // 配置初始化（加载到内存），和数据库配置加载。数据库是“持久化记忆”，内存是“工作记忆”。

	// Initialize token store
	channel.TokenStoreInit()  // 管理具有有效期的第三方令牌（Token）

	// Initialize HTTP server
	server := gin.Default()
	server.SetHTMLTemplate(common.LoadTemplate())
	//server.Use(gzip.Gzip(gzip.DefaultCompression))  // conflict with sse

	// Initialize session store
	var store sessions.Store
	if common.RedisEnabled {  // Redis模式
		opt := common.ParseRedisOption()
		store, _ = redis.NewStore(opt.MinIdleConns, opt.Network, opt.Addr, opt.Password, []byte(common.SessionSecret))
	} else {
		store = cookie.NewStore([]byte(common.SessionSecret))
	}
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   30 * 24 * 3600,
	})
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server, buildFS, indexPage)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}
	err = server.Run(":" + port)
	if err != nil {
		log.Println(err)
	}
}

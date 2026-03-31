package middleware

import (
	"encoding/json"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"message-pusher/common"
	"net/http"
	"net/url"
)

type turnstileCheckResponse struct {
	Success bool `json:"success"`
}

func TurnstileCheck() gin.HandlerFunc {  // 在执行高风险操作（如注册、登录、发验证码）之前，验证请求是真实人类发出的，还是**自动化脚本（机器人）**发出的。（未开启）
	return func(c *gin.Context) {
		if common.TurnstileCheckEnabled {  // 1. 检查全局配置：是否开启了人机验证
			session := sessions.Default(c)  // 2. 获取当前请求的 Session
			turnstileChecked := session.Get("turnstile")  // 3. 优化体验：检查 Session 中是否记录过“已验证成功”
			if turnstileChecked != nil {
				c.Next()
				return
			}
			response := c.Query("turnstile")  // 4. 获取前端提交的验证令牌（通常从 URL 参数中获取）
			if response == "" {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Turnstile token 为空",
				})
				c.Abort()
				return
			}
			rawRes, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", url.Values{
				"secret":   {common.TurnstileSecretKey},
				"response": {response},
				"remoteip": {c.ClientIP()},
			})
			if err != nil {
				common.SysError(err.Error())
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				c.Abort()
				return
			}
			defer rawRes.Body.Close()
			var res turnstileCheckResponse
			err = json.NewDecoder(rawRes.Body).Decode(&res)
			if err != nil {
				common.SysError(err.Error())
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				c.Abort()
				return
			}
			if !res.Success {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Turnstile 校验失败，请刷新重试！",
				})
				c.Abort()
				return
			}
			session.Set("turnstile", true)
			err = session.Save()
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message": "无法保存会话信息，请重试",
					"success": false,
				})
				return
			}
		}
		c.Next()
	}
}

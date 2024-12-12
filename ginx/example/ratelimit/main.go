package main

import (
	"github.com/DaHuangQwQ/gpkg/ginx"
	"github.com/DaHuangQwQ/gpkg/ginx/middleware/ratelimit"
	"github.com/DaHuangQwQ/gpkg/logger"
	ratelimiter "github.com/DaHuangQwQ/gpkg/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

func main() {
	server := ginx.NewServer()
	client := redis.NewClient(nil)

	server.Use(ratelimit.NewBuilder(
		ratelimiter.NewRedisSlidingWindowLimiter(client, time.Second, 10), logger.NewNoOpLogger()).Build())

	server.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, ginx.Result{
			Code: 0,
			Msg:  "ok",
			Data: "hello world",
		})
	})
	_ = server.Start(":8081")
}

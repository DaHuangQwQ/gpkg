package main

import (
	"github.com/DaHuangQwQ/gpkg/ginx"
	"github.com/DaHuangQwQ/gpkg/ginx/middleware/prometheus"
	"github.com/gin-gonic/gin"
)

func main() {
	server := ginx.NewServer()

	builder := prometheus.Builder{
		Namespace:  "test",
		Subsystem:  "test",
		Name:       "user",
		InstanceId: "1",
		Help:       "1",
	}
	server.Use(builder.BuildActiveRequest())
	server.Use(builder.BuildResponseTime())

	server.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, ginx.Result{
			Code: 0,
			Msg:  "ok",
			Data: "hello world",
		})
	})
	_ = server.Start(":8081")
}

package main

import (
	"fmt"
	"github.com/DaHuangQwQ/gpkg/ginx"
	ijwt "github.com/DaHuangQwQ/gpkg/ginx/jwt"
	"github.com/DaHuangQwQ/gpkg/ginx/middleware/jwt_token"
	"github.com/gin-gonic/gin"
	"time"
)

func main() {
	server := ginx.NewServer()
	jwtHandler := ijwt.NewLocalJWTHandler(
		[]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"),
		[]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixA"),
		time.Minute*30)
	server.Use(jwt_token.NewBuilder(jwtHandler).IgnorePaths("/login").Build())
	server.GET("/login", func(ctx *gin.Context) {
		_ = jwtHandler.SetLoginToken(ctx, 1)
	})
	server.GET("/", func(ctx *gin.Context) {
		claims := ctx.MustGet("claims").(ijwt.UserClaims)
		ctx.JSON(200, ginx.Result{
			Code: 0,
			Msg:  "ok",
			Data: fmt.Sprintf("hello world %d", claims.Uid),
		})
	})
	_ = server.Start(":8081")
}

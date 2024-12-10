package ginx

import (
	"github.com/DaHuangQwQ/gpkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

var L logger.Logger

func NewWarpLogger(l logger.Logger) {
	L = l
}

func WarpWithToken[Req any](fn func(ctx *gin.Context, req Req, u UserClaims) (Result, error)) (string, gin.HandlerFunc) {
	var (
		path string
		req  Req
	)
	t := reflect.TypeOf(req)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Meta" {
			path = field.Tag.Get("path")
		}
	}
	DocGen(req)
	return path, func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			ctx.JSON(http.StatusOK, Result{
				Code: 5,
				Msg:  "参数错误" + err.Error(),
			})
			return
		}
		res := ctx.MustGet("claims")
		if res == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		user, ok := res.(UserClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		result, err := fn(ctx, req, user)
		if err != nil {
			ctx.JSON(http.StatusOK, result)
			L.Info("系统错误", logger.Field{Key: "err", Val: err})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func Warp[Req any](fn func(ctx *gin.Context, req Req) (Result, error)) (string, gin.HandlerFunc) {
	var (
		path string
		req  Req
	)
	t := reflect.TypeOf(req)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Meta" {
			path = field.Tag.Get("path")
		}
	}
	DocGen(req)
	return path, func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			ctx.JSON(http.StatusOK, Result{
				Code: 5,
				Msg:  "参数错误" + err.Error(),
			})
			return
		}

		result, err := fn(ctx, req)
		if err != nil {
			ctx.JSON(http.StatusOK, result)
			L.Info("系统错误", logger.Field{Key: "err", Val: err})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Wrap(fn func(c *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := fn(ctx)
		if err != nil {
			ctx.JSON(http.StatusOK, result)
			L.Info("系统错误", logger.Field{Key: "err", Val: err})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

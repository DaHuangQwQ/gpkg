package ginx

import (
	"github.com/DaHuangQwQ/gpkg/ginx/openapi"
	"github.com/DaHuangQwQ/gpkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

var (
	L   logger.Logger
	Oai = openapi.NewOpenAPI()
)

func NewWarpLogger(l logger.Logger) {
	L = l
}

func WrapWithToken[Req any, Res any](fn func(ctx *gin.Context, req Req, u UserClaims) (Result[Res], error)) (string, string, gin.HandlerFunc) {
	var (
		method string
		path   string
		req    Req
	)
	t := reflect.TypeOf(req)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Meta" {
			path = field.Tag.Get("path")
			method = field.Tag.Get("method")
		}
	}

	route := openapi.Route[Res, Req]{
		Operation:            nil,
		FullName:             "",
		Path:                 path,
		AcceptedContentTypes: nil,
		DefaultStatusCode:    0,
		Method:               method,
		Middlewares:          nil,
	}
	err := route.RegisterOpenAPIOperation(Oai)
	if err != nil {
		panic(err)
	}

	return method, path, func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			ctx.JSON(http.StatusOK, Result[Res]{
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

func Wrap[Req any, Res any](fn func(ctx *gin.Context, req Req) (Result[Res], error)) (string, string, gin.HandlerFunc) {
	var (
		method string
		path   string
		req    Req
	)
	t := reflect.TypeOf(req)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Meta" {
			path = field.Tag.Get("path")
			method = field.Tag.Get("method")
		}
	}

	route := openapi.Route[Res, Req]{
		Operation:            nil,
		FullName:             "",
		Path:                 path,
		AcceptedContentTypes: nil,
		DefaultStatusCode:    0,
		Method:               method,
		Middlewares:          nil,
	}
	err := route.RegisterOpenAPIOperation(Oai)
	if err != nil {
		panic(err)
	}

	return method, path, func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			ctx.JSON(http.StatusOK, Result[Res]{
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

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type Meta struct{}

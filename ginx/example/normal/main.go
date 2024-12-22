package main

import (
	"github.com/DaHuangQwQ/gpkg/ginx"
	"github.com/gin-gonic/gin"
)

type UserGetReq struct {
	ginx.Meta `method:"GET" path:"users/:id"`
	Id        int `json:"id"`
}

func getUser(ctx *gin.Context, req UserGetReq) (ginx.Result, error) {
	return ginx.Result{
		Code: 0,
		Msg:  "ok",
		Data: "hello",
	}, nil
}

func main() {
	server := ginx.NewServer()
	server.Handle(ginx.Wrap[UserGetReq](getUser))
	_ = server.Start(":8080")
}

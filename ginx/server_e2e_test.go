//go:build e2e

package ginx

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestServer(t *testing.T) {
	server := NewServer()
	server.Handle(Warp[UserGetReq](getUser))
	_ = server.Start(":8080")
}

func getUser(ctx *gin.Context, req UserGetReq) (Result, error) {
	return Result{
		Code: 0,
		Msg:  "ok",
		Data: "nihao",
	}, nil
}

type UserGetReq struct {
	Meta `method:"GET" path:"users/:id"`
	Id   int `json:"id"`
}

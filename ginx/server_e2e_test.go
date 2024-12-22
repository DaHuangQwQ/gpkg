//go:build e2e

package ginx

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"testing"
)

func TestServer(t *testing.T) {
	server := NewServer()
	server.Handle(Wrap[userGetReq, userGetRes](getUser))

	marshal, err := json.Marshal(Oai.Description())
	if err != nil {
		return
	}
	println(string(marshal))
}

func getUser(ctx *gin.Context, req userGetReq) (Result[userGetRes], error) {
	return Result[userGetRes]{
		Code: 0,
		Msg:  "ok",
		Data: userGetRes{},
	}, nil
}

type userGetReq struct {
	Meta `method:"GET" path:"users/:id"`
	Id   int `json:"id" validate:"required,min=1,max=32"`
}

type userGetRes struct {
	Meta
}

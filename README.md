# go pkg
提供常用工具的封装
```shell
go get github.com/DaHuangQwQ/gpkg
```

## gin
1. api文档生成
2. jwt中间件
3. 限流中间件
4. 可观测中间件
5. 简化代码
```go
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
	server := ginx.NewServer(":8080")
	server.Handle(ginx.Warp[UserGetReq](getUser))
	_ = server.Start()
}

```
## gorm
1. 可观测中间件
2. 双写
3. 读写分离
4. 分库分表
## grpc
1. 负载均衡算法
2. 日志中间件
3. 普罗米修斯
4. 限流
5. 熔断
6. 链路追踪
7. 路由策略
```go
package main

import (
    "context"
    "github.com/DaHuangQwQ/gpkg/grpcx"
    "github.com/DaHuangQwQ/gpkg/grpcx/registry/etcd"
    clientv3 "go.etcd.io/etcd/client/v3"
    "time"
)

// client
func main() {
	etcdClient, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:12379"},
	})

	r, _ := etcd.NewRegistry(etcdClient)

	client, _ := grpcx.NewClient(grpcx.ClientWithRegistry(r, time.Second*3))

	_, _ = client.Dial(context.Background(), "user_service")
}

// server
func main() {
    etcdClient, _ := clientv3.New(clientv3.Config{
    Endpoints: []string{"localhost:12379"},
    })
    
    r, _ := etcd.NewRegistry(etcdClient)
    
    server := grpcx.NewServer("user_service", grpcx.WithRegistry(r))
    
    _ = server.Start(":8082")
}
```
## logger
简化代码
## migrator
不停机数据迁移方案
- 全量修复
- 增量修复
## net
获取本机ip
## ratelimit
- 滑动窗口算法
## redis
- 可观测中间件
## sarama
- 简化代码
- kafka分批处理
## app
- 简化代码
## canal
1. 定义统一接口
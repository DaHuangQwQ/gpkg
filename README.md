# go pkg
提供常用工具的封装
```shell
go get github.com/DaHuangQwQ/gpkg
```

## ginx

移至其他仓库
https://github.com/DaHuangQwQ/ginx

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
8. 分组路由
9. 组播
10. 广播
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
## ratelimit
- 滑动窗口算法
- 固定窗口算法
- 令牌桶算法
- 漏桶算法
## migrator
不停机数据迁移方案
- 全量修复
- 增量修复
## redis
- 可观测中间件
## sarama
kafka 消息队列
- 简化代码
- kafka分批处理
- 延时队列
## app
- 简化代码
## canal
1. 定义统一接口
## logger
简化代码
## net
获取本机ip
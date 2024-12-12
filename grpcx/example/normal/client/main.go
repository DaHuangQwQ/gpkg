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

package main

import (
	"github.com/DaHuangQwQ/gpkg/grpcx"
	"github.com/DaHuangQwQ/gpkg/grpcx/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// server
func main() {
	etcdClient, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:12379"},
	})

	r, _ := etcd.NewRegistry(etcdClient)

	server := grpcx.NewServer("user_service", grpcx.WithRegistry(r))

	_ = server.Start(":8082")
}

package grpcx

import (
	"context"
	"github.com/DaHuangQwQ/gpkg/grpcx/registry/etcd"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestClient1(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	client, err := NewClient(ClientWithRegistry(r, time.Second*3), ClientWithBalancer("custom_wrr"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err = client.Dial(ctx, "user_service")
	require.NoError(t, err)
}

func TestServer1(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	server := NewServer("user_service", WithRegistry(r), WithWeight(100))

	err = server.Start(":8082")
	require.NoError(t, err)
}

func TestServer2(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	server := NewServer("user_service", WithRegistry(r), WithWeight(120))

	err = server.Start(":8083")
	require.NoError(t, err)
}

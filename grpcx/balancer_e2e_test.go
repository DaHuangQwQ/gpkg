package grpcx

import (
	"context"
	"fmt"
	"github.com/DaHuangQwQ/gpkg/grpcx/balancer/round_robin"
	"github.com/DaHuangQwQ/gpkg/grpcx/registry/etcd"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
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
	for i := range 3 {
		go func(i int) {
			server := NewServer("user_service", WithRegistry(r), WithWeight(100))

			err = server.Start(fmt.Sprintf("localhost:808%d", i+2))
			require.NoError(t, err)
			t.Logf("server%d start", i)
		}(i)
	}

	time.Sleep(time.Second * 2)

	client, err := NewClient(ClientWithRegistry(r, time.Second*3), ClientWithBalancer("round_robin",
		&round_robin.PickerBuilder{
			Filter: func(info balancer.PickInfo, addr resolver.Address) bool {
				return true
			},
		}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	conn, err := client.Dial(ctx, "user_service")
	require.NoError(t, err)
	t.Log(conn)
}

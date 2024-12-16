package grpcx

import (
	"context"
	"fmt"
	_ "github.com/DaHuangQwQ/gpkg/grpcx/balancer/round_robin"
	_ "github.com/DaHuangQwQ/gpkg/grpcx/balancer/wrr"
	"github.com/DaHuangQwQ/gpkg/grpcx/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"time"
)

type ClientOption func(c *Client)

type Client struct {
	insecure bool
	r        registry.Registry
	timeout  time.Duration

	balancer balancer.Builder
}

func NewClient(opts ...ClientOption) (*Client, error) {
	res := &Client{
		insecure: true,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func (c *Client) Dial(ctx context.Context, service string, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.r != nil {
		rb, err := NewRegistryBuilder(c.r, c.timeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithResolvers(rb))
	}
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	if len(dialOptions) > 0 {
		opts = append(opts, dialOptions...)
	}
	if c.balancer != nil {
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, c.balancer.Name())))
	}

	return grpc.DialContext(ctx, fmt.Sprintf("registry:///%s", service), opts...)
}

func ClientInsecure() ClientOption {
	return func(c *Client) {
		c.insecure = true
	}
}
func ClientWithRegistry(r registry.Registry, timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.r = r
		c.timeout = timeout
	}
}

func ClientWithBalancer(name string, pb base.PickerBuilder) ClientOption {
	return func(c *Client) {
		c.balancer = base.NewBalancerBuilder(name, pb, base.Config{HealthCheck: true})
		balancer.Register(c.balancer)
	}
}

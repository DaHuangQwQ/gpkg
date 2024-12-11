package grpcx

import (
	"context"
	"github.com/DaHuangQwQ/gpkg/grpcx/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOption func(server *Server)

type Server struct {
	*grpc.Server
	r    registry.Registry
	Name string

	registerTimeout time.Duration

	// 负载均衡
	weight uint32
}

func NewServer(name string, opts ...ServerOption) *Server {
	res := &Server{
		Server:          grpc.NewServer(),
		Name:            name,
		registerTimeout: time.Second * 3,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

// Start 启动服务器并且阻塞
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	if s.r != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
		defer cancel()
		err = s.r.Register(ctx, registry.ServiceInstance{
			Name:    s.Name,
			Address: addr,
			Weight:  s.weight,
		})
		if err != nil {
			return err
		}
	}

	return s.Serve(listener)
}

func (s *Server) Close() error {
	if s.r != nil {
		_ = s.r.Close()
	}
	s.Server.GracefulStop()
	return nil
}

func WithRegistry(r registry.Registry) ServerOption {
	return func(server *Server) {
		server.r = r
	}
}

func WithWeight(w uint32) ServerOption {
	return func(server *Server) {
		server.weight = w
	}
}

package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

const name = "round_robin"

func init() {
	balancer.Register(base.NewBalancerBuilder(name, &PickerBuilder{}, base.Config{HealthCheck: true}))
}

type PickerBuilder struct {
}

func (b *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for conn := range info.ReadySCs {
		connections = append(connections, conn)
	}
	return &Balancer{
		index:       0,
		connections: connections,
	}
}

type Balancer struct {
	index       uint32
	connections []balancer.SubConn
}

// Pick 负载均衡轮询算法
func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	subConn := b.connections[b.index]
	atomic.AddUint32(&b.index, 1)
	return balancer.PickResult{
		SubConn: subConn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

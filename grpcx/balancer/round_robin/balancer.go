package round_robin

import (
	balance "github.com/DaHuangQwQ/gpkg/grpcx/balancer"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync/atomic"
)

const name = "round_robin"

func init() {
	balancer.Register(base.NewBalancerBuilder(name, &PickerBuilder{
		Filter: func(info balancer.PickInfo, addr resolver.Address) bool {
			return true
		},
	}, base.Config{HealthCheck: true}))
}

type PickerBuilder struct {
	Filter balance.Filter
}

func (b *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]subConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		connections = append(connections, subConn{
			c:    conn,
			addr: connInfo.Address,
		})
	}

	return &Balancer{
		index:       0,
		connections: connections,
		filter:      b.Filter,
	}
}

type Balancer struct {
	index       uint32
	connections []subConn
	filter      balance.Filter
}

// Pick 负载均衡轮询算法
func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]subConn, 0, len(b.connections))
	// group filter
	for _, conn := range b.connections {
		if b.filter == nil || !b.filter(info, conn.addr) {
			continue
		}
		candidates = append(candidates, conn)
	}

	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	conn := candidates[int(b.index)%len(candidates)]
	atomic.AddUint32(&b.index, 1)
	return balancer.PickResult{
		SubConn: conn.c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type subConn struct {
	c    balancer.SubConn
	addr resolver.Address
}

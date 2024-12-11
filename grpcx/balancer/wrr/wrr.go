package wrr

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"math"
	"strconv"
	"sync"
)

const name = "custom_wrr"

func init() {
	// NewBalancerBuilder 帮我们 PickerBuilder 转化为 BalancerBuilder
	balancer.Register(base.NewBalancerBuilder("custom_wrr", &PickerBuilder{}, base.Config{HealthCheck: true}))
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	conns := make([]*conn, 0, len(info.ReadySCs))
	for subConn, connInfo := range info.ReadySCs {
		weightStr := connInfo.Address.Attributes.Value("weight").(string)
		weight, err := strconv.ParseInt(weightStr, 10, 64)
		if err != nil {
			weight = 1
		}

		conns = append(conns, &conn{
			weight:        uint32(weight),
			currentWeight: uint32(weight),
			cc:            subConn,
			available:     true,
		})
	}
	return &Picker{
		conns: conns,
	}
}

// Picker 平滑的加权轮询算法
type Picker struct {
	conns []*conn
	mutex sync.Mutex
}

// Pick 基于权重的负载均衡算法
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var (
		totalWeight uint32
		maxCC       *conn
	)

	for _, connInfo := range p.conns {
		if connInfo.available == false {
			continue
		}
		totalWeight += connInfo.weight
		connInfo.currentWeight += connInfo.weight
		if maxCC == nil || maxCC.currentWeight < connInfo.currentWeight {
			maxCC = connInfo
		}
	}

	maxCC.currentWeight -= totalWeight

	return balancer.PickResult{
		SubConn: maxCC.cc,
		Done: func(info balancer.DoneInfo) {
			// 很多动态算法 根据结果来 调整权重
			maxCC.mutex.Lock()
			if info.Err != nil && maxCC.currentWeight == 0 {
				return
			}
			if info.Err == nil && maxCC.currentWeight == math.MaxUint32 {
				return
			}
			if info.Err != nil {
				maxCC.currentWeight--
			} else {
				maxCC.currentWeight++
			}
			maxCC.mutex.Unlock()

			switch info.Err {
			case context.Canceled:
				return
			case context.DeadlineExceeded:
				return
			case io.EOF, io.ErrUnexpectedEOF:
				// 节点已经崩了
				maxCC.available = false
				return
			default:
				st, ok := status.FromError(info.Err)
				if ok {
					code := st.Code()
					switch code {
					case codes.Unavailable:
						// 这里可能表达的是 熔断
						// 挪走该节点， 该节点已经不可用
						maxCC.available = false
						go func() {
							// 开一个额外的 goroutine 去探活
							// 借助 health check
							// for loop
							if p.healthCheck(maxCC) {
								maxCC.available = true
								// 最好加点流量控制的措施
								// maxCC.currentWeight
								// 掷骰子
							}
						}()
					case codes.ResourceExhausted:
						// 这里可能表达的是 限流
						// 可以挪走 可以留着，留着把两个权重一起调低

						// 加一个错误码 表示降级
					}
				}
			}
		},
	}, nil
}

func (p *Picker) healthCheck(cc *conn) bool {
	// 调用 GRPC 内置的 healthCheck 接口
	return true
}

type conn struct {
	mutex         sync.Mutex
	weight        uint32
	currentWeight uint32
	cc            balancer.SubConn
	available     bool
	// vip 节点 非 vip 节点 VIP节点全崩了 考虑挤占非VIP 节点
	group string
}

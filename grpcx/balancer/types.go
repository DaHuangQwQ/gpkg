package balancer

import "google.golang.org/grpc/balancer"

type Filter func(info balancer.PickInfo) bool

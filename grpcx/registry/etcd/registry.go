package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DaHuangQwQ/gpkg/grpcx/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync"
)

type Registry struct {
	etcd    *clientv3.Client
	session *concurrency.Session

	cancels []func()
	mutex   sync.Mutex

	close chan struct{}
}

func NewRegistry(etcd *clientv3.Client) (*Registry, error) {
	session, err := concurrency.NewSession(etcd)
	if err != nil {
		return nil, err
	}
	return &Registry{
		etcd:    etcd,
		session: session,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.ServiceInstance) error {
	val, err := json.Marshal(si)
	if err != nil {
		return err
	}
	_, err = r.etcd.Put(ctx, r.instanceKey(si), string(val), clientv3.WithLease(r.session.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.ServiceInstance) error {
	_, err := r.etcd.Delete(ctx, r.instanceKey(si))
	return err
}

func (r *Registry) ListServices(ctx context.Context, serviceName string) ([]registry.ServiceInstance, error) {
	getResp, err := r.etcd.Get(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make([]registry.ServiceInstance, 0, len(getResp.Kvs))
	for _, kv := range getResp.Kvs {
		var si registry.ServiceInstance
		if err := json.Unmarshal(kv.Value, &si); err != nil {
			return nil, err
		}
		res = append(res, si)
	}
	return res, nil
}

func (r *Registry) Subscribe(serviceName string) <-chan registry.Event {
	ctx, cancel := context.WithCancel(context.Background())

	r.mutex.Lock()
	r.cancels = append(r.cancels, cancel)
	r.mutex.Unlock()

	watch := r.etcd.Watch(clientv3.WithRequireLeader(ctx), r.serviceKey(serviceName), clientv3.WithPrefix())

	res := make(chan registry.Event)

	go func() {
		for {
			select {
			case resp := <-watch:
				if resp.Err() != nil {
					continue
				}
				if resp.Canceled {
					return
				}
				for range resp.Events {
					res <- registry.Event{}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return res
}

func (r *Registry) Close() error {
	r.mutex.Lock()
	cancels := r.cancels
	r.cancels = nil
	r.mutex.Unlock()
	close(r.close)
	for _, cancel := range cancels {
		cancel()
	}
	return nil
}

func (r *Registry) instanceKey(si registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", si.Name, si.Address)
}

func (r *Registry) serviceKey(sn string) string {
	return fmt.Sprintf("/micro/%s", sn)
}

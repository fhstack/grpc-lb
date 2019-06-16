package etcdv3

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/v3/clientv3"
	"log"
	"net"
	"strings"
	"time"
)

var Deregister = make(chan struct{})

//Register to register
func Register(target, service, host, port string, usettl bool, ttl int) error {
	serviceValue := net.JoinHostPort(host, port)
	serviceKey := fmt.Sprintf("/%s/%s/%s", schema, service, serviceValue)

	var err error
	//etcd address
	config := clientv3.Config{
		Endpoints:            strings.Split(target, ","),
		AutoSyncInterval:     0,
		DialTimeout:          time.Second,
		DialKeepAliveTime:    time.Second,
		DialKeepAliveTimeout: time.Second,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  nil,
		Username:             "",
		Password:             "",
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
		LogConfig:            nil,
		PermitWithoutStream:  false,
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return fmt.Errorf("grpclb: create etcdv3 failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	if usettl {
		resp, err := cli.Grant(ctx, int64(ttl))
		cancel()
		if err != nil {
			return fmt.Errorf("gprclb: create etcdv3 lease failed: %v", err)
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*1)
		_, err = cli.Put(ctx, serviceKey, serviceValue, clientv3.WithLease(resp.ID))
		cancel()
		if err != nil {
			return fmt.Errorf("grpclb: set service '%s' with ttl to etcdv3 failed: %s", service, err.Error())
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*1)
		_, err = cli.KeepAlive(ctx, resp.ID)
		cancel()
		if err != nil {
			return fmt.Errorf("grpclb: refresh service '%s' with ttl to etcdv3 failed: %s", service, err.Error())
		}
	} else {
		_, err = cli.Put(ctx, serviceKey, serviceValue)
		cancel()
		if err != nil {
			return fmt.Errorf("grpclb: set service '%s' to etcdv3 failed: %s", service, err.Error())
		}
	}

	go func() {
		<-Deregister
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*1)
		_, err := cli.Delete(ctx, serviceKey)
		cancel()
		if err != nil {
			log.Printf("delete service %s failed: %v", service, err)
		}
		Deregister <- struct{}{}

	}()

	return nil
}

func UnRegister() {
	Deregister <- struct{}{}
	<-Deregister
}

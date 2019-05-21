package main

import (
	"context"
	"flag"
	grpclb "github.com/l-f-h/grpc-lb/etcdv3"
	"github.com/l-f-h/grpc-lb/testPack/pb_generate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"log"
	"strconv"
	"time"
)

var (
	service  = flag.String("service", "HelloService", "service name")
	register = flag.String("register", "127.0.0.1:2379", "register etcd address")
)

func main() {
	flag.Parse()
	r := grpclb.NewResolver(*register, *service)
	resolver.Register(r)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	conn, err := grpc.DialContext(ctx, r.Scheme()+"://authority/"+*service, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	cancel()
	if err != nil {
		log.Printf("gprc Dial service: '%s' failed: %v", *service, err)
		return
	}
	log.Println("already")
	ticker := time.NewTicker(time.Second * 1)
	for t := range ticker.C {
		client := pb_gen.NewGreeterClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		resp, err := client.SayHello(ctx, &pb_gen.HelloRequest{Name: "liufeihao " + strconv.Itoa(t.Second())})
		cancel()
		if err != nil {
			log.Printf("Request SayHello failed: %v", err)
			return
		}
		log.Printf("Response: %s", resp.Resp)
	}
}

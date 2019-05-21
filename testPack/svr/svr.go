package main

import (
	"context"
	"flag"
	grpclb "github.com/l-f-h/grpc-lb/etcdv3"
	"github.com/l-f-h/grpc-lb/testPack/pb_generate"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	service  = flag.String("service", "HelloService", "service name")
	host     = flag.String("host", "127.0.0.1", "listening host")
	port     = flag.String("port", "50001", "listening port")
	register = flag.String("register", "127.0.0.1:2379", "register etcd address")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", net.JoinHostPort(*host, *port))
	if err != nil {
		log.Fatalf("TCP listen failed: %v", err)
	}

	//150 just a test magic number
	err = grpclb.Register(*register, *service, *host, *port, 150)
	if err != nil {
		log.Fatalf("Register service:'%s' failed: %v", *service, err)
	}
	log.Println("Register successfully")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		log.Printf("reveive signal '%v'", s)
		grpclb.UnRegister()
		os.Exit(1)
	}()

	log.Printf("starting %s at %s:%s", *service, *host, *port)
	s := grpc.NewServer()
	pb_gen.RegisterGreeterServer(s, &server{})
	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("Service '%s' started failed: %v", *service, err)
	}
}

type server struct{}

func (s *server) SayHello(ctx context.Context, req *pb_gen.HelloRequest) (*pb_gen.HelloResponse, error) {
	log.Printf("Receive %s", req.Name)
	return &pb_gen.HelloResponse{
		Resp: "Hello " + req.Name + " from " + net.JoinHostPort(*host, *port),
	}, nil
}

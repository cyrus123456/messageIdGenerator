package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":8028")
	if err != nil {
		log.Println("tcp链接异常\n\r")
	}
	log.Println("消息id生成器服务tcp端口:8028\n\r")
	grpcServer := grpc.NewServer()
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Println("grpcServer错误\n\r")
	}
}

package main

import (
	"context"
	"log"
	"messageIdGenerator/getMessageIGrpcGateway"
	"net"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Server struct {
	getMessageIGrpcGateway.UnimplementedGetMessageIdServiceServer
}

func (_this *Server) GetMessageId(ctx context.Context, request *getMessageIGrpcGateway.GetMessageRequest) (*getMessageIGrpcGateway.GetMessageResponse, error) {
	log.Println("GetMessageId接收到参数\n\r", request)
	return &getMessageIGrpcGateway.GetMessageResponse{
		UserMessageId: "572387598302",
	}, nil
}

func main() {

	wg := sync.WaitGroup{}

	wg.Add(2)

	go grpcServer(&wg)

	go httpGetWay(&wg)

	wg.Wait()

}

func httpGetWay(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	//	http getway ******************************************************************
	gwmux := runtime.NewServeMux()

	//Http
	gwServer := &http.Server{
		Handler: gwmux,
		Addr:    "127.0.0.1:8080",
	}

	//转发服务
	grpcClientConn, err := grpc.DialContext(
		context.Background(),
		"127.0.0.1:8028",
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)

	if err != nil {
		log.Println("getWay Http 转发异常\n\r", err)
	} else {
		log.Println("getWay Http 转发正常端口8080\n\r")
	}
	defer grpcClientConn.Close()

	// 合并处理
	err = getMessageIGrpcGateway.RegisterGetMessageIdServiceHandler(
		context.Background(),
		gwmux,
		grpcClientConn,
	)

	if err != nil {
		log.Println("getWay Http 处理异常\n\r", err)
	} else {
		log.Println("getWay Http 处理正常端口\n\r")
	}

	err = gwServer.ListenAndServe()
	if err != nil {
		log.Println("getWay Http 服务异常\n\r", err)
	} else {
		log.Println("getWay Http 服务正常端口8080\n\r")
	}

}

func grpcServer(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	//	grpc 服务 ******************************************************************

	listener, err := net.Listen("tcp", "127.0.0.1:8028")
	if err != nil {
		log.Println("tcp链接异常\n\r", err)
	} else {
		log.Println("tcp链接正常端口8028\n\r")
	}

	grpcServer := grpc.NewServer()
	//注册服务
	getMessageIGrpcGateway.RegisterGetMessageIdServiceServer(grpcServer, &Server{})
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Println("grpcServer错误\n\r", err)
	} else {
		log.Println("grpcServer正常\n\r")
	}
}

package main

import (
	"context"
	"log"
	"messageIdGenerator/dao"
	"messageIdGenerator/getMessageIGrpcGateway"
	"net"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type ServerCLASS struct {
	getMessageIGrpcGateway.UnimplementedGetMessageIdServiceServer
}

func (_this *ServerCLASS) GetMessageId(
	ctx context.Context,
	request *getMessageIGrpcGateway.GetMessageRequest,
) (*getMessageIGrpcGateway.GetMessageResponse, error) {

	log.Println(
		"GetMessageId接收到参数\n\r",
		request,
	)

	// val, err := dao.Redisdb.HMSet(
	// 	ctx,
	// 	request.UserId,
	// 	map[string]interface{}{
	// 		"CurId":        "0",
	// 		"MaxId":        "10000",
	// 		"MaxIdsEction": "9457",
	// 	},
	// ).Result()

	// if err != nil {
	// 	log.Println("用户", request.UserId, "缓存失败\n\r", err)
	// } else {
	// 	log.Println("用户", request.UserId, "缓存的值\n\r", val)
	// }

	type SECTION_ID_ROW struct {
		sectionId    int `json:"sectionId,omitempty"`
		sectionMaxId int `json:"sectionMaxId,omitempty"`
	}

	var sectionIdRow SECTION_ID_ROW

	// 非常重要：确保QueryRow之后调用Scan方法，否则持有的数据库链接不会被释放 [注意点]

	sqlStr := "select * from text.mid where sectionId = ?"

	log.Println("mysql开始查值")
	row := dao.MysqlDbConn.QueryRow(sqlStr, 1)
	log.Println("mysql查询一行的结果", *row, row)

	err := row.Scan(&sectionIdRow.sectionId, &sectionIdRow.sectionMaxId)
	if err != nil {
		log.Println("mysql取值失败", err)
	} else {
		log.Println("mysql取值成功", sectionIdRow)
	}

	return &getMessageIGrpcGateway.GetMessageResponse{
		UserMessageId: "572387598302",
	}, nil
}

func main() {
	defer dao.MysqlDbConn.Close()
	defer dao.Redisdb.Close()

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
		Addr:    "127.0.0.1:8034",
	}

	//转发服务
	grpcClientConn, err := grpc.DialContext(
		context.Background(),
		"127.0.0.1:8056",
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

	listener, err := net.Listen("tcp", "127.0.0.1:8056")
	if err != nil {
		log.Println("tcp链接异常\n\r", err)
	} else {
		log.Println("tcp链接正常端口8056\n\r")
	}

	grpcServer := grpc.NewServer()
	//注册服务
	getMessageIGrpcGateway.RegisterGetMessageIdServiceServer(grpcServer, &ServerCLASS{})
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Println("grpcServer错误\n\r", err)
	} else {
		log.Println("grpcServer正常\n\r")
	}
}

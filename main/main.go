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

//http接口处理函数
func (_this *ServerCLASS) GetMessageId(
	ctx context.Context,
	request *getMessageIGrpcGateway.GetMessageRequest,
) (*getMessageIGrpcGateway.GetMessageResponse, error) {

	// 请求入参	UserId  SectionId
	log.Println("GetMessageId接收到参数", request)

	// 操作redis*********************************************************************
	// 数据库结构 schame key:UserId ; values: CurId MaxId step
	fields, err := dao.Redisdb.HMGet(ctx, request.UserId, "CurId", "MaxId", "step", "belongingUserNum").Result()
	if err != nil {
		type SECTION_ID_ROW struct {
			sectionId        int
			id               int
			sectionMaxId     int
			belongingUserNum int
			step             int
		}

		var sectionIdRow SECTION_ID_ROW

		log.Println("redis查询失败👺👺", err)
		// 如果获取不到查看myqsl

		// 操作mysql*********************************************************************

		// 非常重要：确保QueryRow之后调用Scan方法，否则持有的数据库链接不会被释放 [注意点]

		sqlStr := "select * from text.usertosectionid where id = ?"

		log.Println("mysql开始查值")
		row := dao.MysqlDbConn.QueryRow(sqlStr, request.UserId)
		log.Println("mysql查询一行的结果", *row, row)

		err = row.Scan(&sectionIdRow.sectionId)
		if err != nil {
			log.Println("mysql取值失败👺", err)
			// 1 获取id最新数据，归属用户+1，
			// 2 如果没有即新增一条数据，归属用户+1

			sqlStr := "select max(id) from text.mid"
			row := dao.MysqlDbConn.QueryRow(sqlStr)
			log.Println("mysql查询一行的结果", *row, row)
			err = row.Scan(
				&sectionIdRow.id,
				&sectionIdRow.sectionMaxId,
				&sectionIdRow.belongingUserNum,
				&sectionIdRow.step,
			)
			if err != nil {
				//mysql 查询失败👺，拆入一条数据,并且UI对应表也插入
				// id，sectionMaxId，belongingUserNum，step
				sqlStr = "insert into text.mid() values ()"
				ret, err := dao.MysqlDbConn.Exec(sqlStr)
				if err != nil {
					log.Printf("插入 失败👺, err:%v", err)
				}
				// 新插入数据的id
				theID, err := ret.LastInsertId()
				if err != nil {
					log.Printf("get lastinsert ID failed, err:%v", err)
				}
				log.Printf("插入 成功, the id is %d.", theID)

				sqlStr = "insert into text.usertosectionid(sectionId) values (?)"
				ret, err = dao.MysqlDbConn.Exec(sqlStr, theID)
				if err != nil {
					log.Printf("插入 失败👺, err:%v", err)
				}
				// 新插入数据的id
				theID, err = ret.LastInsertId()
				if err != nil {
					log.Printf("get lastinsert ID failed, err:%v", err)
				}
				log.Printf("插入 成功, the id is %d.", theID)
			}

			type SECTION_ID_ROW struct {
				sectionMaxId     int
				step             int
				belongingUserNum int
			}

			var sectionIdRow SECTION_ID_ROW
			sqlStr = "select sectionMaxId,step,belongingUserNum from text.mid join text.usertosectionid on (text.usertosectionid.sectionId = text.mid.id)"
			row = dao.MysqlDbConn.QueryRow(sqlStr)
			log.Println("mysql查询一行的结果", *row, row)
			err = row.Scan(
				&sectionIdRow.sectionMaxId,
				&sectionIdRow.belongingUserNum,
				&sectionIdRow.step,
			)
			if err != nil {
				log.Println("mysql关联查询失败👺", err)
			} else {
				log.Println("mysql关联查询成功", sectionIdRow)
			}

			// 数据库结构 schame key:UserId ; values: CurId MaxId step belongingUserNum
			ok, err := dao.Redisdb.HMSet(
				ctx,
				request.UserId,
				map[string]interface{}{
					"CurId":            "1",
					"MaxId":            sectionIdRow.sectionMaxId,
					"step":             sectionIdRow.step,
					"belongingUserNum": "1",
				},
			).Result()
			if err != nil {
				log.Println("存redis失败👺", err)
			} else {
				log.Println("存redis成功", ok)
				// 查询redis
				fields, err = dao.Redisdb.HMGet(ctx, request.UserId, "CurId", "MaxId", "step", "belongingUserNum").Result()
				if err != nil {
					log.Println("最终获取redis失败👺", err)
				} else {
					log.Println("最终获取redis成功", fields)
				}
			}

		} else {
			log.Println("mysql取值成功", sectionIdRow)
		}
	} else {
		log.Println("redis查询成功", fields)
	}

	return &getMessageIGrpcGateway.GetMessageResponse{
		UserMessageId: "572387598302",
	}, nil
}

func main() {
	log.SetFlags(log.Lshortfile)

	defer dao.MysqlDbConn.Close()
	defer dao.Redisdb.Close()

	wg := sync.WaitGroup{}

	wg.Add(2)

	// grpc服务
	go grpcServer(&wg)

	// http转发代理服务
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
		Addr:    "127.0.0.1:4589",
	}

	//转发服务
	grpcClientConn, err := grpc.DialContext(
		context.Background(),
		"127.0.0.1:9876",
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)

	if err != nil {
		log.Println("getWay Http 转发异常", err)
	} else {
		log.Println("getWay Http 转发正常端口8080")
	}
	defer grpcClientConn.Close()

	// 合并处理
	err = getMessageIGrpcGateway.RegisterGetMessageIdServiceHandler(
		context.Background(),
		gwmux,
		grpcClientConn,
	)

	if err != nil {
		log.Println("getWay Http 处理异常", err)
	} else {
		log.Println("getWay Http 处理正常端口")
	}

	err = gwServer.ListenAndServe()
	if err != nil {
		log.Println("getWay Http 服务异常", err)
	} else {
		log.Println("getWay Http 服务正常端口8080")
	}

}

func grpcServer(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	//	grpc 服务 ******************************************************************

	listener, err := net.Listen("tcp", "127.0.0.1:9876")
	if err != nil {
		log.Println("tcp链接异常", err)
	} else {
		log.Println("tcp链接正常端口9876")
	}

	grpcServer := grpc.NewServer()
	//注册服务
	getMessageIGrpcGateway.RegisterGetMessageIdServiceServer(grpcServer, &ServerCLASS{})
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Println("grpcServer错误", err)
	} else {
		log.Println("grpcServer正常")
	}
}

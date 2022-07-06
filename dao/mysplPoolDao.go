package dao

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var MysqlDbConn *sql.DB

func init() {

	// 数据源语法："用户名:密码@[连接方式](主机名:端口号)/数据库名"
	MysqlDb, err := sql.Open(
		"mysql",
		"root:123456@tcp(127.0.0.1:3306)/text",
	)
	if err != nil {
		log.Println("mysql数据库链接失败", err)
	} else {
		log.Println("mysql数据库链接成功")
	}
	// 要写到err下面
	// defer MysqlDb.Close()

	MysqlDbConn = MysqlDb

	// 尝试与数据库建立连接（校验DSN是否正确）
	err = MysqlDbConn.Ping()
	if err != nil {
		log.Println("mysql数据库ping失败", err)
	} else {
		log.Println("mysql数据库ping成功")
	}

	MysqlDb.SetConnMaxLifetime(time.Minute * 3)
	// 设置与数据库建立连接的最大数目
	MysqlDb.SetMaxOpenConns(1024)
	// 设置连接池中的最大闲置连接数，0 表示不会保留闲置。
	MysqlDb.SetMaxIdleConns(0)
	log.Println("数据库初始化连接成功!")
}

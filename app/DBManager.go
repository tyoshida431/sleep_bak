package main

import (
	//"database/sql"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

// DBに接続するための構造体
type Server struct {
	DB *sqlx.DB
}

func dbConnect() (*sqlx.DB, error) {
	var err error
	// .envファイルを読み込む
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error open .env file:", err)
		return nil, err
	}

	// 環境変数を変数に格納する
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbAddr := os.Getenv("DB_ADDRESS")
	dbName := os.Getenv("DB_NAME")

	// 接続プロパティをキャプチャする
	config := mysql.Config{
		User:                 dbUser,
		Passwd:               dbPass,
		Net:                  "tcp",
		Addr:                 dbAddr,
		DBName:               dbName,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	// データベースを開く
	db, err := sqlx.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatalf("DB open Error:%v", err)
		return nil, err
	}

	// 接続が有効であるか確認する
	err = db.Ping()
	if err != nil {
		log.Fatalf("DB ping Error:%v", err)
		return nil, err
	}
	return db, nil
}

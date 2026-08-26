package main

import (
	//"log"
	"github.com/gin-gonic/gin"
)

func main() {
	// Ginのルーターを作成
	router := gin.Default()
	cors, err := getCorsConfig()
	if err == nil {
		router.Use(cors)
	}

	// ルートエンドポイントを定義し、Hello Worldを返す
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	// ポート7070でサーバーを起動
	router.Run(":7070")
}

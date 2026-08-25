package main

import (
  "github.com/gin-gonic/gin"
)

func main() {
  // Ginのルーターを作成
  router := gin.Default()

  // ルートエンドポイントを定義し、Hello Worldを返す
  router.GET("/", func(c *gin.Context) {
    c.JSON(200, gin.H{
    "message": "Hello World",
    })
  })

  // ポート7070でサーバーを起動
  router.Run(":7070")
}

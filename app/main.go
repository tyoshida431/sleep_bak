package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	var err error
	router := gin.Default()
	cors, err := getCorsConfig()
	if err == nil {
		router.Use(cors)
	}
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})
	router.GET("/sleep", func(c *gin.Context) {
		month := c.Query("month")
		data, err := getSleep(month)
		if err == nil {
			c.JSON(200, data)
		} else {
			c.JSON(500, gin.H{"message": "サーバー内部でエラーが発生しました"})
		}
	})
	router.POST("/sleep", func(c *gin.Context) {
		//var sleeps []SleepFromFront
		//if err := c.ShouldBindJSON(&sleeps); err != nil {
		//	c.JSON(500, gin.H{"message": "リクエストが不正です。"})
		//}
		//data, err := updateSleep(sleeps)
		//if err == nil {
		//	c.JSON(200, data)
		//} else {
		//	c.JSON(500, gin.H{"message": "サーバー内部でエラーが発生しました"})
		//}
	})
	err = router.Run(":7070")
	if err != nil {
		log.Fatal("Cannot Run Server Error:", err)
	}
}

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
	} else {
		log.Fatal("Get Cors Config Error: ", err)
	}
	router.GET("/sleep", func(c *gin.Context) {
		month := c.Query("month")
		sleeps, err := getSleep(month)
		if err == nil {
			c.JSON(200, sleeps)
		} else {
			c.JSON(500, gin.H{"message": err})
		}
	})
	router.POST("/sleep", func(c *gin.Context) {
		var sleeps []SleepFromFront
		if err := c.ShouldBindJSON(&sleeps); err != nil {
			c.JSON(500, gin.H{"message": err})
		} else {
			sleeps, err := updateSleep(sleeps)
			if err == nil {
				c.JSON(200, sleeps)
			} else {
				c.JSON(500, gin.H{"message": err})
			}
		}
	})
	if err = router.Run(":7070"); err != nil {
		log.Fatal("Cannot Run Server Error: ", err)
	}
}

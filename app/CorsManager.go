package main

import (
	"encoding/json"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
)

type URL struct {
	Url string `json:"url"`
}

type Config struct {
	APICorsAllowOrigins []URL `json:"APICorsAllowOrigins"`
}

func loadAllowOrigins() ([]string, error) {
	var retUrls []string
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal("loadConfig os.Open Error:", err)
		return nil, err
	}
	defer file.Close()
	jsonData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("ioutil.ReadAll Error:", err)
		return nil, err
	}
	var conf Config
	json.Unmarshal(jsonData, &conf)
	for _, url := range conf.APICorsAllowOrigins {
		retUrls = append(retUrls, url.Url)
	}
	return retUrls, nil
}

func getCorsConfig() (gin.HandlerFunc, error) {
	var allowOrigins []string
	allowOrigins, err := loadAllowOrigins()
	if err != nil {
		log.Fatal("load CorsOrigins Error:", err)
		return nil, err
	}
	config := cors.DefaultConfig()
	config.AllowOrigins = allowOrigins
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Access-Control-Allow-Headers", "X-CSRF-Token", "Access-Control-Allow-Origin"}
	config.AllowMethods = []string{"GET", "POST"}
	return cors.New(config), nil
}

package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	defer func() {
		fileCloseErr := file.Close()
		if fileCloseErr != nil {
			log.Fatal("File Close Error:", err)
			return
		}
		err = fileCloseErr
	}()
	jsonData, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("ioutil.ReadAll Error:", err)
		return nil, err
	}
	var conf Config
	err = json.Unmarshal(jsonData, &conf)
	if err != nil {
		log.Fatal("json Unmarshal Error:", err)
		return nil, err
	}
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

package main

import(
  "io/ioutil"
  "log"
  "os"
  "encoding/json"
  "github.com/gin-gonic/gin"
  "github.com/gin-contrib/cors"
)

type URL struct{
  Url string `json:"url"`
}

type Config struct{
  APICorsAllowOrigins []URL `json:"APICorsAllowOrigins"`
}

func loadConfig() ([]string, error){
  var ret []string
  file,err:=os.Open("config.json")
  if err!=nil{
    log.Fatal("loadConfig os.Open err:",err)
    return nil,err
  }
  defer file.Close()
  jsonData,err:=ioutil.ReadAll(file)
  if err!=nil{
    log.Fatal("ioutil.ReadAll err:",err)
    return nil,err
  }
  var conf Config
  json.Unmarshal(jsonData,&conf)
  for _,url:=range conf.APICorsAllowOrigins{
    ret=append(ret,url.Url)
  }
  return ret,nil
}

func corsMiddleware()(gin.HandlerFunc,error){
  var err error
  var allowOrigins []string

  confList,err:=loadConfig()
  if err!=nil{
    log.Fatal("load CorsOrigins Error:",err)
    return nil,err
  }
  for _,conf:=range confList{
    allowOrigins=append(allowOrigins,conf)
  }
  
  config:=cors.DefaultConfig()
  config.AllowOrigins=allowOrigins
  config.AllowHeaders=[]string{"Origin","Content-Type","Accept","Access-Control-Allow-Headers","X-CSRF-Token","Access-Control-Allow-Origin",}
  config.AllowMethods=[]string{"GET","POST"}
  return cors.New(config),nil
}

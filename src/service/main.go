package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"service/handler"
	"service/module"
	"service/conf"
)

func main() {
	flag.Parse()

	module.SessionTable = make(map[int]*module.Session)

	if glog.V(conf.LOG_LV) {
		glog.Info("wechat-service start")
	}

	route := gin.Default()

	route.POST("/api/create", handler.SessionCreate)
	route.POST("/api/login", handler.LoginScan)
	route.POST("/api/send", handler.SendMessage)
	route.POST("/api/exit", handler.Exit)

	route.Run(":8888")
}

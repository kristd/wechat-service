package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"service/conf"
	"service/handler"
	"service/module"
)

func main() {
	flag.Parse()
	module.SessionTable = make(map[int]*module.Session)

	gin.SetMode(gin.DebugMode)
	route := gin.Default()

	route.POST(conf.API_CREATE, handler.SessionCreate)
	route.POST(conf.API_LOGIN, handler.LoginScan)
	route.POST(conf.API_SEND, handler.SendMessage)
	route.POST(conf.API_EXIT, handler.Exit)

	go module.StatNotify()

	route.Run(conf.PORT)
}

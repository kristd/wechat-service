package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

var SessionTable map[int]*Session

func main() {
	flag.Parse()
	SessionTable = make(map[int]*Session)

	if glog.V(2) {
		glog.Info("Service Start")
	}

	route := gin.Default()

	route.POST("/api/create", SessionCreate)
	route.POST("/api/login", LoginScan)
	route.POST("/api/send", SendMessage)
	route.POST("/api/exit", Exit)

	route.Run(":8888")
}

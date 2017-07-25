package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"service/common"
	"service/conf"
	"service/handler"
	"service/module"
	"time"
)

var (
	cfgFile string
)

func init() {
	module.SessionTable = make(map[int]*module.Session)
	flag.StringVar(&cfgFile, "config", "cfg.json", "config file")
}

func StatNotify() {
	for {
		userIDs := make([]int, 0)

		for _, v := range module.SessionTable {
			if v.GetSrvStatus() {
				userIDs = append(userIDs, v.UserID)
			}
		}

		data := &common.Msg_Heartbeat_Request{
			Action:  conf.CLIENT_BEAT,
			UserIDs: userIDs,
		}

		b, _ := json.Marshal(data)
		body := bytes.NewReader(b)
		req, _ := http.NewRequest("POST", conf.Config.NOTIFY_ADDR, body)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		glog.Info(">>> [StatNotify] Heartbeat Start = [", req, "]")

		clinet := &http.Client{}
		resp, err := clinet.Do(req)
		if err != nil {
			glog.Error("[StatNotify] Clinet.Do err = ", err)
		} else {
			glog.Info(">>> [StatNotify] Heartbeat End")
		}
		defer resp.Body.Close()

		time.Sleep(conf.HEARTBEAT_INTERVAL * time.Second)
	}
}

func main() {
	flag.Parse()

	if !conf.LoadConfig(cfgFile) {
		return
	}

	gin.SetMode(gin.DebugMode)
	route := gin.Default()

	route.POST(conf.API_CREATE, handler.SessionCreate)
	route.POST(conf.API_LOGIN, handler.LoginScan)
	route.POST(conf.API_SEND, handler.SendMessage)
	route.POST(conf.API_EXIT, handler.Exit)

	if conf.Config.NOTIFY_ON {
		go StatNotify()
	}

	route.Run(conf.Config.PORT)
}

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"os"
	"os/signal"
	"service/common"
	"service/conf"
	"service/handler"
	"service/module"
	"time"
	"syscall"
)

var (
	cfgFile string
	quit	chan os.Signal
)

func init() {
	quit = make(chan os.Signal)
	module.SessionTable = make(map[int]*module.Session)

	flag.StringVar(&cfgFile, "config", "cfg.json", "Config file")
}

func StatNotify() {
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
	resp.Body.Close()
}

func Stop(q chan os.Signal)  {
	signal.Notify(q, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	<-q

	for _, v := range module.SessionTable {
		v.Stop()
	}

	if conf.Config.NOTIFY_ON {
		StatNotify()
	}

	if conf.Config.MONGODB != "" {
		module.DisConnect()
		glog.Info(">>> [Stop] Disconnect from mongodb")
	}

	glog.Info(">>> [Stop] Service exit ", len(module.SessionTable))
	os.Exit(0)
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

	if conf.Config.MONGODB != "" {
		m := module.GetDBInstant()
		if m == nil {
			return
		}
	}

	if conf.Config.NOTIFY_ON {
		go func() {
			for {
				StatNotify()
				time.Sleep(conf.HEARTBEAT_INTERVAL * time.Second)
			}
		}()
	}

	go Stop(quit)
	route.Run(conf.Config.PORT)
}

package module

import (
	"bytes"
	"encoding/json"
	"github.com/golang/glog"
	"net/http"
	"service/common"
	"service/conf"
	"time"
)

type HeartBeat struct {
}

func StatNotify() {
	for {
		userIDs := make([]int, 0)

		for _, v := range SessionTable {
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
		req, _ := http.NewRequest(conf.HEARTBEAT_METHOD, conf.HEARTBEAT_ADDR, body)
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

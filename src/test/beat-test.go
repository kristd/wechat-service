package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"service/common"
	"service/conf"
)

func main() {
	userIDs := []int{1, 2, 3}

	data := &common.Msg_Heartbeat_Request{
		Action:  conf.CLIENT_BEAT,
		UserIDs: userIDs,
	}

	b, _ := json.Marshal(data)
	body := bytes.NewReader(b)
	req, _ := http.NewRequest("POST", "http://120.92.234.2:7652/api/go/wechat/online", body)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	clinet := &http.Client{}
	resp, err := clinet.Do(req)
	if err != nil {
		fmt.Println("[StatNotify] Clinet.Do err = ", err)
	}
	defer resp.Body.Close()

	//time.Sleep(conf.HEARTBEAT_INTERVAL * time.Second)
}

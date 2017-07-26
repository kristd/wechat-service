package main

import (
    "service/common"
    "service/conf"
    "encoding/json"
    "bytes"
    "net/http"
    "fmt"
)

func main() {
    userIDs := []int {6,13,17}

    data := &common.Msg_Heartbeat_Request{
        Action:  conf.CLIENT_BEAT,
        UserIDs: userIDs,
    }

    b, _ := json.Marshal(data)
    body := bytes.NewReader(b)
    req, _ := http.NewRequest("POST", "http://120.92.52.210:9502/api/go/wechat/online", body)
    req.Header.Set("Content-Type", "application/json; charset=utf-8")

    clinet := &http.Client{}
    resp, err := clinet.Do(req)
    if err != nil {
        fmt.Println("Clinet.Do err = ", err)
    }
    defer resp.Body.Close()
}
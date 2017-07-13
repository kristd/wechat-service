package main

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	//"golang.org/x/net/websocket"
	"io/ioutil"
	"net/http"
	"strings"
)


func main() {
	data := "{\"action\":1,\"id\":6,\"conf\":[{\"group\":\"广东麻将\",\"keywords\":[{\"keyword\":\"游戏\",\"cotent\":\"http://qyq.xoyo.com/h5/download/?app_id=XYd0ogCwfB4wYCqdikYooVe\",\"img\":\"/data/pic_tmp/bf3/510/a68/1d4f42e0cdcb0aaf8c8f35c5d297c2ab.gif\"}]}]}"

	//http
	body := strings.NewReader(data)

	req, err := http.NewRequest("POST", "http://127.0.0.1:8888/api/create", body)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	fmt.Println("req =", req)

	clinet := &http.Client{}
	resp, err := clinet.Do(req)
	if err != nil {
		fmt.Println("clinet.Do err =", err)
	}

	defer resp.Body.Close()
	ret, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(ret))
}

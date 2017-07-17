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
	data := "{\"action\":1,\"id\":6,\"conf\":[{\"group\":\"测试\",\"keywords\":[{\"keyword\":\"麻将\",\"text\":\"http://qyq.xoyo.com/h5/download/?app_id=XYd0ogCwfB4wYCqdikYooVe\",\"img\":\"\"}]}]}"

	body := strings.NewReader(data)

	req, err := http.NewRequest("POST", "http://localhost:8888/api/create", body)
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

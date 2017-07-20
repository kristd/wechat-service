package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	data := "{\"action\":4,\"id\":2,\"group\":\"测试\",\"params\":{\"type\":2,\"method\":\"new_player\",\"content\":\"http://weixin.xoyo.com/award/images/logo.png\"}}"

	body := strings.NewReader(data)
	//req, err := http.NewRequest("POST", "http://120.92.234.72:8888/api/send", body)
	req, err := http.NewRequest("POST", "http://localhost:8888/api/send", body)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	//fmt.Println("req =", req)

	clinet := &http.Client{}
	resp, err := clinet.Do(req)
	if err != nil {
		fmt.Println("clinet.Do err =", err)
	}

	defer resp.Body.Close()

	ret, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(ret))
}

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	data := "{\"action\":4,\"id\":6,\"group\":\"测试\",\"params\":{\"type\":1,\"method\":\"new_player\",\"content\":\"悦己刚刚进入广东麻将[24328]号房，1缺3，大家快来打牌。\"}}"

	body := strings.NewReader(data)
	req, err := http.NewRequest("POST", "http://localhost:8888/api/send", body)
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

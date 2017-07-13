package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	IMG_PATH = "/Users/kristd/Documents/sublime/image/"
)

func getImg(url string) (n int64, err error) {
	path := strings.Split(url, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	}
	name = IMG_PATH + name
	fmt.Println(name)
	out, err := os.Create(name)
	defer out.Close()
	resp, err := http.Get(url)
	defer resp.Body.Close()
	pix, err := ioutil.ReadAll(resp.Body)
	n, err = io.Copy(out, bytes.NewReader(pix))
	return n, err
}

func main() {
	//getImg("http://weixin.xoyo.com/award/images/logo.png")
	data := "{\"action\":4,\"id\":6,\"group\":\"测试\",\"params\":{\"type\":2,\"method\":\"new_player\",\"content\":\"http://weixin.xoyo.com/award/images/logo.png\"}}"

	body := strings.NewReader(data)
	req, err := http.NewRequest("POST", "http://127.0.0.1:8888/api/login", body)
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

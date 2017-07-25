package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func hello(resp http.ResponseWriter, req *http.Request) {
	b, _ := ioutil.ReadAll(req.Body)
	fmt.Println("body =", string(b))
	return
}

func main() {
	http.HandleFunc("/api/create", hello)

	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		fmt.Println("err = ", err)
	}
}

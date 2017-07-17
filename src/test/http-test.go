package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	fmt.Println("body =", string(b))
	return
}

func main() {
	http.HandleFunc("/", hello)
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		fmt.Println("err = ", err)
	}
}

package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
)

func HandleConn(conn *websocket.Conn) {
	body := make([]byte, 1024)
	n, _ := conn.Read(body)
	fmt.Println("body = [", string(body[:n]), "]")
}

func main() {
	http.Handle("/websocket", websocket.Handler(HandleConn))
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		fmt.Println("serve err = [", err, "]")
	}
}

package main

import (
	"golang.org/x/net/websocket"
	"log"
)

func main() {
	ws, err := websocket.Dial("ws://localhost:8888/websocket", "", "http://localhost:8888/websocket")
	if err != nil {
		log.Fatal("err = ", err)
	} else {
		log.Println("success")
	}

	ws.Write([]byte("HelloWorld"))
	ws.Close()
}

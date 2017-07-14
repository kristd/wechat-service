package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"
)


func main() {
    data := "{\"action\":5,\"id\":6}"

    body := strings.NewReader(data)
    req, err := http.NewRequest("POST", "http://127.0.0.1:8888/api/exit", body)
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

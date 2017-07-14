package main

import (
    "github.com/gin-gonic/gin"
    "github.com/golang/glog"
    "encoding/json"
    "net/http"
    "fmt"
)

func makeExitResponse(uid, code int, msg string) *Msg_Exit_Response {
    resp := &Msg_Exit_Response{
        Action: 1,
        UserID: uid,
        Code:   code,
        Msg:    msg,
    }
    return resp
}

func Exit(c *gin.Context) {
    exit_request := &Msg_Exit_Request{}
    exit_response := &Msg_Exit_Response{}

    reqMsgBuf := make([]byte, MAX_BUF_SIZE)

    n, _ := c.Request.Body.Read(reqMsgBuf)
    if glog.V(2) {
        glog.Info(">>>>> Request Body <<<<<", string(reqMsgBuf[:n]))
    }

    err := json.Unmarshal(reqMsgBuf[:n], exit_request)
    if err != nil {
        glog.Info("Client_Action_Exit Unmarshal err ", err)
        exit_response = makeExitResponse(0, 10001, "")
        c.JSON(http.StatusBadRequest, exit_response)
        return
    } else {
        glog.Info("Client_Action_Exit exit_request ", exit_request)
    }

    s, succ := SessionTable[exit_request.UserID]
    if !succ {
        fmt.Println("Get session failed")
    }

    s.Stop()
    delete(SessionTable, exit_request.UserID)

    repsMsg := &Msg_Exit_Response{
        Action: Client_Action_Exit,
        UserID: s.UserID,
    }

    c.JSON(http.StatusOK, repsMsg)

    return
}

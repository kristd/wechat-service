package main

import (
    "github.com/gin-gonic/gin"
    "fmt"
    "encoding/json"
    "net/http"
)

func makeSendResponse(uid, code int, msg string) *Msg_Send_Response {
    resp := &Msg_Send_Response{
        Action: 2,
        UserID: uid,
        Code:   code,
        Msg:    msg,
    }
    return resp
}

func SendMessage(c *gin.Context)  {
    send_request := &Msg_Send_Request{}
    send_response := &Msg_Send_Response{}

    reqMsgBuf := make([]byte, MAX_BUF_SIZE)

    n, _ := c.Request.Body.Read(reqMsgBuf)

    err := json.Unmarshal(reqMsgBuf[:n], send_request)
    if err != nil {
        fmt.Println("Client_Action_Send err1 =", err)
    }

    s, ok := SessionTable[send_request.UserID]
    if !ok {
        fmt.Println("Session not exit")
        return
    }

    //ToUsers := s.Cm.GetContactByName(reqMsg.Group)
    toUser := &User{}
    for _, toUser = range s.Cm.cl {
        fmt.Println(">>>>> Users Nick Name <<<<<", toUser.NickName)
        fmt.Println(">>>>> Users User Name <<<<<", toUser.UserName)

        if toUser.NickName == send_request.Group {
            break
        }
    }

    var msgID string
    var localID string
    var result string

    switch send_request.Params.Type {
    case TEXT_MSG:
        msgID, localID, err = s.SendText(send_request.Params.Content, s.Bot.UserName, toUser.UserName)
        if msgID != "" && localID != "" {
            fmt.Println("send msg succecss")
            result = "success"
        } else {
            fmt.Println("SendText err =", err)
            result = "failed"
        }
    case IMG_MSG:
        fileName, err := LoadImage(send_request.Params.Content)
        if err != nil {
            fmt.Println("LoadImage err = ", err)
            result = "failed"
        } else {
            retcd, err := s.SendImage(fileName, s.Bot.UserName, toUser.UserName)
            if retcd == 0 {
                fmt.Println("send image succecss")
                result = "success"
            } else if err != nil {
                fmt.Println("SendImage err =", err)
                result = "failed"
            }
        }
    default:
        result = "failed"
    }

    fmt.Println(result)

    resp := &Msg_Send_Response{
        Action: Client_Action_Login,
        UserID: s.UserID,
        Code:   200,
        Msg:    "success",
    }

    respBuf, err := json.Marshal(resp)
    if err != nil {
        fmt.Println("Client_Action_Send err2 =", err, respBuf)
    }

    send_response = makeSendResponse(s.UserID, 200, "")
    c.JSON(http.StatusOK, send_response)
}
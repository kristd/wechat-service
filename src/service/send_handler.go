package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func makeSendResponse(uid, code int, msg string) *Msg_Send_Response {
	resp := &Msg_Send_Response{
		Action: CLIENT_SEND,
		UserID: uid,
		Code:   code,
		Msg:    msg,
	}
	return resp
}

func SendMessage(c *gin.Context) {
	send_request := &Msg_Send_Request{}
	send_response := &Msg_Send_Response{}

	reqMsgBuf := make([]byte, MAX_BUF_SIZE)

	n, _ := c.Request.Body.Read(reqMsgBuf)

	err := json.Unmarshal(reqMsgBuf[:n], send_request)
	if err != nil {
		fmt.Println("CLIENT_SEND err1 =", err)
	}

	s, ok := SessionTable[send_request.UserID]
	if !ok {
		fmt.Println("Session not exit")
		return
	}

	//ToUsers := s.contactMgr.GetContactByName(reqMsg.Group)
	toUser := &User{}
	for _, toUser = range s.contactMgr.contactList {
		fmt.Println(">>>>> Users Nick Name <<<<<", toUser.NickName)
		fmt.Println(">>>>> Users User Name <<<<<", toUser.UserName)

		if toUser.NickName == send_request.Group {
			break
		}
	}

	var msgID string
	var localID string

	switch send_request.Params.Type {
	case TEXT_MSG:
		msgID, localID, err = s.SendText(send_request.Params.Content, s.bot.UserName, toUser.UserName)
		if msgID != "" && localID != "" {
			fmt.Println("send msg success")
		} else {
			fmt.Println("SendText err =", err)
		}
	case IMG_MSG:
		//fileName, err := LoadImage(send_request.Params.Content)
		fileName := "/Users/kristd/Documents/sublime/image/logo.png"
		if err != nil {
			fmt.Println("LoadImage err = ", err)
		} else {
			retcd, err := s.SendImage(fileName, s.bot.UserName, toUser.UserName)
			if retcd == 0 {
				fmt.Println("send image success")
			} else if err != nil {
				fmt.Println("SendImage err =", err)
			}
		}
	}

	send_response = makeSendResponse(s.userID, 200, "success")

	c.JSON(http.StatusOK, send_response)
	return
}

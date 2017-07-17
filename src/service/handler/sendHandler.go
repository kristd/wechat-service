package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"service/conf"
	"service/common"
	"service/module"
)

func makeSendResponse(uid, code int, msg string) *common.Msg_Send_Response {
	resp := &common.Msg_Send_Response{
		Action: conf.CLIENT_SEND,
		UserID: uid,
		Code:   code,
		Msg:    msg,
	}
	return resp
}

func SendMessage(c *gin.Context) {
	send_request := &common.Msg_Send_Request{}
	send_response := &common.Msg_Send_Response{}

	reqMsgBuf := make([]byte, conf.MAX_BUF_SIZE)

	n, _ := c.Request.Body.Read(reqMsgBuf)

	err := json.Unmarshal(reqMsgBuf[:n], send_request)
	if err != nil {
		fmt.Println("CLIENT_SEND err1 =", err)
	}

	s, ok := module.SessionTable[send_request.UserID]
	if !ok {
		fmt.Println("Session not exit")
		return
	}

	//ToUsers := s.contactMgr.GetContactByName(reqMsg.Group)
	toUser := &common.User{}
	for _, toUser = range s.ContactMgr.ContactList {
		fmt.Println(">>>>> Users Nick Name <<<<<", toUser.NickName)
		fmt.Println(">>>>> Users User Name <<<<<", toUser.UserName)

		if toUser.NickName == send_request.Group {
			break
		}
	}

	var msgID string
	var localID string

	switch send_request.Params.Type {
	case conf.TEXT_MSG:
		msgID, localID, err = s.SendText(send_request.Params.Content, s.Bot.UserName, toUser.UserName)
		if msgID != "" && localID != "" {
			fmt.Println("send msg success")
		} else {
			fmt.Println("SendText err =", err)
		}
	case conf.IMG_MSG:
		//fileName, err := LoadImage(send_request.Params.Content)
		fileName := "./image/logo.png"
		if err != nil {
			fmt.Println("LoadImage err = ", err)
		} else {
			retcd, err := s.SendImage(fileName, s.Bot.UserName, toUser.UserName)
			if retcd == 0 {
				fmt.Println("send image success")
			} else if err != nil {
				fmt.Println("SendImage err =", err)
			}
		}
	}

	send_response = makeSendResponse(s.UserID, 200, "success")

	c.JSON(http.StatusOK, send_response)
	return
}

package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"io"
	"net/http"
	"service/common"
	"service/conf"
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
	//n, _ := c.Request.Body.Read(reqMsgBuf)
	n, _ := io.ReadFull(c.Request.Body, reqMsgBuf)

	err := json.Unmarshal(reqMsgBuf[:n], send_request)
	if err != nil {
		glog.Error("[SendMessage] request json data unmarshal err = [", err, "]")
		send_response = makeSendResponse(send_request.UserID, -30000, "request json format error")
	} else {
		glog.Info(">>> [SendMessage] Request JSON Data = [", send_request, "]")
		s, exist := module.SessionTable[send_request.UserID]
		if !exist {
			glog.Error("[SendMessage] UserID = ", send_request.UserID, " session not exist")
			send_response = makeSendResponse(send_request.UserID, -30001, "Session not exist")
		} else {
			toUser := &common.User{
				NickName: "",
				UserName: "",
			}

			if send_request.UserType == conf.USER_GROUP {
				for _, u := range s.ContactMgr.GetGroupContacts() {
					if u.NickName == send_request.NickName {
						toUser = u
						break
					}
				}
			} else {
				for _, u := range s.ContactMgr.GetPersonContacts() {
					if u.NickName == send_request.NickName {
						toUser = u
						break
					}
				}
			}

			if toUser.UserName == "" {
				glog.Error("[SendMessage] User ", send_request.UserID, " group [", send_request.NickName, "] not found")
				send_response = makeSendResponse(send_request.UserID, -30002, "group not found")
			} else {
				switch send_request.Params.Type {
				case conf.TEXT_MSG:
					s.SendText(send_request.Params.Content, s.Bot.UserName, toUser.UserName)
				case conf.IMG_MSG:
					s.SendImage(send_request.Params.Content, s.Bot.UserName, toUser.UserName)
				}
				send_response = makeSendResponse(s.UserID, 200, "success")
			}
		}
	}

	c.JSON(http.StatusOK, send_response)
	return
}

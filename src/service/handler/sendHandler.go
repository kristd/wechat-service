package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"service/conf"
	"service/common"
	"service/module"
	"github.com/golang/glog"
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
		if glog.V(2) {
			glog.Error("[SendMessage] request json data unmarshal err = [", err, "]")
		}
		send_response = makeSendResponse(send_request.UserID, -30000, "request json format error")
	} else {
		s, exist := module.SessionTable[send_request.UserID]
		if !exist {
			if glog.V(2) {
				glog.Error("[SendMessage] request json data unmarshal err = [", err, "]")
			}
			send_response = makeSendResponse(send_request.UserID, -30001, "request json format error")
		} else {
			toUser := &common.User{
				NickName:	"",
				UserName:	"",
			}

			for _, u := range s.ContactMgr.ContactList {
				if u.NickName == send_request.Group {
					toUser = u
					break
				}
			}

			if toUser.UserName == "" {
				if glog.V(2) {
					glog.Error("[SendMessage] User ", send_request.UserID, " group ", send_request.Group, " not found")
				}
				send_response = makeSendResponse(send_request.UserID, -30002, "group not found")
			} else {
				switch send_request.Params.Type {
				case conf.TEXT_MSG:
					msgID, localID, err := s.SendText(send_request.Params.Content, s.Bot.UserName, toUser.UserName)
					if msgID != "" && localID != "" {
						if glog.V(2) {
							glog.Info(">>> [SendMessage] User ", send_request.UserID, " send text message success")
						}
						send_response = makeSendResponse(s.UserID, 200, "success")
					} else {
						if glog.V(2) {
							glog.Error("[SendMessage] User ", send_request.UserID, " send text message failed, err = [", err, "]")
						}
						send_response = makeSendResponse(s.UserID, -30003, "send text message failed")
					}
				case conf.IMG_MSG:
					//fileName, err := utils.LoadImage(send_request.Params.Content)
					fileName := "./image/logo.png"
					if err != nil {
						if glog.V(2) {
							glog.Error("[SendMessage] User ", send_request.UserID, " load image message failed, err = [", err, "]")
						}
						send_response = makeSendResponse(s.UserID, -30004, "load image message failed")
					} else {
						retcd, err := s.SendImage(fileName, s.Bot.UserName, toUser.UserName)
						if retcd == 0 {
							if glog.V(2) {
								glog.Info(">>> [SendMessage] User ", send_request.UserID, " send image message success")
							}
							send_response = makeSendResponse(s.UserID, 200, "success")
						} else if err != nil {
							if glog.V(2) {
								glog.Error("[SendMessage] User ", send_request.UserID, " send image message failed, err = [", err, "]")
							}
							send_response = makeSendResponse(s.UserID, -30005, "send image message failed")
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, send_response)
	return
}

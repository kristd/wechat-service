package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"service/conf"
	"service/common"
	"service/module"
)

func makeLoginResponse(uid, status, code int, msg string) *common.Msg_Login_Response {
	resp := &common.Msg_Login_Response{
		Action: conf.CLIENT_LOGIN,
		UserID: uid,
		Code:   code,
		Msg:    msg,
		Status: status,
	}
	return resp
}

func LoginScan(c *gin.Context) {
	login_request := &common.Msg_Login_Request{}
	login_response := &common.Msg_Login_Response{}

	reqMsgBuf := make([]byte, conf.MAX_BUF_SIZE)

	n, _ := c.Request.Body.Read(reqMsgBuf)

	err := json.Unmarshal(reqMsgBuf[:n], login_request)
	if err != nil {
		if glog.V(2) {
			glog.Error("[LoginScan] request json data unmarshal err = [", err, "]")
		}
		login_response = makeLoginResponse(login_request.UserID, 0, -20000, "request json format error")
	} else {
		if glog.V(2) {
			glog.Info(">>> [LoginScan] request json data = [", login_request, "]")
		}

		s, exist := module.SessionTable[login_request.UserID]
		if !exist {
			if glog.V(2) {
				glog.Error("[LoginScan] User ", login_request.UserID, " session not exist")
			}
			login_response = makeLoginResponse(login_request.UserID, 0, -20001, "user session not exist")
		} else {
			stat := s.LoginPolling()
			if stat == 200 {
				s.AnalizeVersion(s.RedirectUrl)
				ret := s.InitUserCookies(s.RedirectUrl)
				if ret == 0 {
					go s.Serve()
					login_response = makeLoginResponse(login_request.UserID, 300, stat, "success")
				} else {
					if glog.V(2) {
						glog.Error("[LoginScan] User ", login_request.UserID, " cookies init failed")
					}
					login_response = makeLoginResponse(login_request.UserID, 0, -20002, "user cookies init failed")
				}
			} else {
				if glog.V(2) {
					glog.Error("[LoginScan] User", login_request.UserID, " login failed, status =", stat)
				}
				login_response = makeLoginResponse(login_request.UserID, 0, -20003, "user login failed")
			}
		}
	}

	c.JSON(http.StatusOK, login_response)
	return
}

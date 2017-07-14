package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
)

func makeLoginResponse(uid, code, status int, msg string) *Msg_Login_Response {
	resp := &Msg_Login_Response{
		Action: CLIENT_LOGIN,
		UserID: uid,
		Code:   code,
		Msg:    msg,
		Status: status,
	}
	return resp
}

func LoginScan(c *gin.Context) {
	statChan := make(chan int, 1)

	login_request := &Msg_Login_Request{}
	login_response := &Msg_Login_Response{}

	reqMsgBuf := make([]byte, MAX_BUF_SIZE)

	n, _ := c.Request.Body.Read(reqMsgBuf)
	if glog.V(2) {
		glog.Info(">>>>> Request Body <<<<<", string(reqMsgBuf[:n]))
	}

	err := json.Unmarshal(reqMsgBuf[:n], login_request)
	if err != nil {
		glog.Info("CLIENT_CREATE Unmarshal err ", err)
		login_response = makeLoginResponse(0, 10001, 10001, "request json format error")
		c.JSON(http.StatusBadRequest, login_response)
		return
	} else {
		glog.Info("CLIENT_CREATE create_request ", login_request)
	}

	s, ok := SessionTable[login_request.UserID]

	if !ok {
		glog.Info("CLIENT_CREATE create_request ", login_request)
		login_response = makeLoginResponse(0, 10001, 10001, "request json format error")
		c.JSON(http.StatusBadRequest, login_response)
		return
	}

	go s.StatusPolling(statChan)

	select {
	case <-statChan:
		statcd := s.GetLoginStat()
		if statcd == 200 {

			fmt.Println("go s.InitAndServe()")

			go s.InitAndServe()
		} else {
			glog.Info(">>> Session status =", statcd)
		}

		login_response = makeLoginResponse(login_request.UserID, 200, statcd, "success")

		c.JSON(http.StatusOK, login_response)
		return
	}
}

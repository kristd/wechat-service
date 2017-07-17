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

func makeLoginResponse(uid, code, status int, msg string) *common.Msg_Login_Response {
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
	statChan := make(chan int, 1)

	login_request := &common.Msg_Login_Request{}
	login_response := &common.Msg_Login_Response{}

	reqMsgBuf := make([]byte, conf.MAX_BUF_SIZE)

	n, _ := c.Request.Body.Read(reqMsgBuf)
	if glog.V(2) {
		glog.Info(">>>>> Request Body <<<<<", string(reqMsgBuf[:n]))
	}

	err := json.Unmarshal(reqMsgBuf[:n], login_request)
	if err != nil {
		glog.Info("CLIENT_LOGIN Unmarshal err ", err)
		login_response = makeLoginResponse(0, 10001, 10001, "request json format error")
		c.JSON(http.StatusBadRequest, login_response)
		return
	} else {
		glog.Info("CLIENT_LOGIN create_request ", login_request)
	}

	s, succ := module.SessionTable[login_request.UserID]

	if !succ {
		glog.Info("CLIENT_CREATE create_request ", login_request)
		login_response = makeLoginResponse(0, 10001, 10001, "request json format error")
		c.JSON(http.StatusBadRequest, login_response)
		return
	}

	go s.StatusPolling(statChan)

	statcd := 0

	select {
	case <-statChan:
		statcd = s.GetLoginStat()
		if statcd == 200 {
			go s.InitAndServe()
		} else {
			glog.Info(">>> Session status =", statcd)
		}
	}

	login_response = makeLoginResponse(login_request.UserID, 200, statcd, "success")

	c.JSON(http.StatusOK, login_response)
	return
}

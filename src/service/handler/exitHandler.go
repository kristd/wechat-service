package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"service/conf"
	"service/common"
	"service/module"
)

func makeExitResponse(uid, code int, msg string) *common.Msg_Exit_Response {
	resp := &common.Msg_Exit_Response{
		Action: conf.CLIENT_EXIT,
		UserID: uid,
		Code:   code,
		Msg:    msg,
	}
	return resp
}

func Exit(c *gin.Context) {
	exit_request := &common.Msg_Exit_Request{}
	exit_response := &common.Msg_Exit_Response{}

	reqMsgBuf := make([]byte, conf.MAX_BUF_SIZE)

	n, _ := c.Request.Body.Read(reqMsgBuf)
	if glog.V(2) {
		glog.Info(">>>>> Request Body <<<<<", string(reqMsgBuf[:n]))
	}

	err := json.Unmarshal(reqMsgBuf[:n], exit_request)
	if err != nil {
		glog.Info("CLIENT_EXIT Unmarshal err ", err)
		exit_response = makeExitResponse(exit_request.UserID, 200, "success")
		c.JSON(http.StatusBadRequest, exit_response)
		return
	} else {
		glog.Info("CLIENT_EXIT exit_request ", exit_request)
	}

	s, succ := module.SessionTable[exit_request.UserID]
	if !succ {
		fmt.Println("Get session failed")
	}

	go s.Stop()

	select {
	case q := <-s.Quit:
		delete(module.SessionTable, exit_request.UserID)
		fmt.Println(">>> Delete Session <<<", q)
	}

	fmt.Println(">>> Stopped")

	exit_response = makeExitResponse(exit_request.UserID, 200, "success")

	c.JSON(http.StatusOK, exit_response)
	return
}

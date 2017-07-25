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
	//n, _ := c.Request.Body.Read(reqMsgBuf)
	n, _ := io.ReadFull(c.Request.Body, reqMsgBuf)

	err := json.Unmarshal(reqMsgBuf[:n], exit_request)
	if err != nil {
		glog.Error("[Exit] request json data unmarshal err = [", err, "]")
		exit_response = makeExitResponse(exit_request.UserID, -40000, "request json data format error")
	} else {
		glog.Info(">>> [Exit] Request JSON Data = [", exit_request, "]")
		s, exist := module.SessionTable[exit_request.UserID]
		if !exist {
			glog.Error("[Exit] User ", exit_request.UserID, " session not exist")
			exit_response = makeExitResponse(exit_request.UserID, -40001, "user session not exist")
		} else {
			s.Stop()
			delete(module.SessionTable, exit_request.UserID)
			//select {
			//case <-s.Quit:
			//	delete(module.SessionTable, exit_request.UserID)
			//}
			glog.Info(">>> [Exit] Session ", exit_request.UserID, " unserve and deleted")
			exit_response = makeExitResponse(exit_request.UserID, 200, "success")
		}
	}

	c.JSON(http.StatusOK, exit_response)
	return
}

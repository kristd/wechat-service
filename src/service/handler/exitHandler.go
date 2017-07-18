package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
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

	n, _ := c.Request.Body.Read(reqMsgBuf)

	err := json.Unmarshal(reqMsgBuf[:n], exit_request)
	if err != nil {
		if glog.V(conf.LOG_LV) {
			glog.Error("[Exit] request json data unmarshal err = [", err, "]")
		}

		exit_response = makeExitResponse(exit_request.UserID, -40000, "request json data format error")
	} else {
		if glog.V(conf.LOG_LV) {
			glog.Info(">>> [Exit] request json data = [", exit_request, "]")
		}

		s, exist := module.SessionTable[exit_request.UserID]
		if !exist {
			if glog.V(conf.LOG_LV) {
				glog.Error("[Exit] User ", exit_request.UserID, " session not exist")
			}

			exit_response = makeExitResponse(exit_request.UserID, -40001, "user session not exist")
		} else {
			go s.Stop()

			select {
			case <-s.Quit:
				delete(module.SessionTable, exit_request.UserID)
			}

			if glog.V(conf.LOG_LV) {
				glog.Info(">>> [Exit] Session ", exit_request.UserID, " unserve and deleted")
			}

			exit_response = makeExitResponse(exit_request.UserID, 200, "success")
		}
	}

	c.JSON(http.StatusOK, exit_response)
	return
}

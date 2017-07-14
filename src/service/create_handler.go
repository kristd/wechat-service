package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"time"
)

func makeCreateResponse(uid int, uuid, qrcode string, code int, msg string) *Msg_Create_Response {
	resp := &Msg_Create_Response{
		Action: 1,
		UserID: uid,
		Code:   code,
		Msg:    msg,
		Uuid:   uuid,
		QrCode: qrcode,
	}
	return resp
}

func SessionCreate(c *gin.Context) {
	create_request := &Msg_Create_Request{}
	create_response := &Msg_Create_Response{}

	reqMsgBuf := make([]byte, MAX_BUF_SIZE)


	n, _ := c.Request.Body.Read(reqMsgBuf)
	if glog.V(2) {
		glog.Info(">>>>> Request Body <<<<<", string(reqMsgBuf[:n]))
	}

	err := json.Unmarshal(reqMsgBuf[:n], create_request)
	if err != nil {
		glog.Info("Client_Action_Create Unmarshal err ", err)
		create_response = makeCreateResponse(0, "", "", 10001, "request json format error")
		c.JSON(http.StatusBadRequest, create_response)
		return
	} else {
		glog.Info("Client_Action_Create create_request ", create_request)
	}

	s := &Session{
		userID:         create_request.UserID,
		wxWebCommon:    DefaultCommon,
		wxWebXcg:       &XmlConfig{},
		wxApi:          &WebwxApi{},
		createTime:     time.Now().Unix(),
		loginStat:      0,
		stop:           false,
        quit:           make(chan bool),
	}

	s.uuID, s.qrcode = s.wxApi.WebwxGetUuid(s.wxWebCommon)

	repsMsg := &Msg_Create_Response{
		Action: Client_Action_Create,
		UserID: s.userID,
		Uuid:   s.uuID,
		QrCode: s.qrcode,
	}

    create_response = makeCreateResponse(s.userID, s.uuID, s.qrcode, 200, "success")
    go s.InitSession(create_request)

    c.JSON(http.StatusOK, repsMsg)
	return
}

package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"service/conf"
	"service/common"
	"service/module"
	"service/wxapi"
	"time"
)

func makeCreateResponse(uid int, uuid, qrcode string, code int, msg string) *common.Msg_Create_Response {
	resp := &common.Msg_Create_Response{
		Action: conf.CLIENT_CREATE,
		UserID: uid,
		Code:   code,
		Msg:    msg,
		Uuid:   uuid,
		QrCode: qrcode,
	}
	return resp
}

func SessionCreate(c *gin.Context) {
	create_request := &common.Msg_Create_Request{}
	create_response := &common.Msg_Create_Response{}

	reqMsgBuf := make([]byte, conf.MAX_BUF_SIZE)

	n, _ := c.Request.Body.Read(reqMsgBuf)

	err := json.Unmarshal(reqMsgBuf[:n], create_request)
	if err != nil {
		if glog.V(2) {
			glog.Error("[SessionCreate] request json data unmarshal err = [", err, "]")
		}

		create_response = makeCreateResponse(create_request.UserID, "", "", -10000, "Request json format error")
	} else {
		if glog.V(2) {
			glog.Info(">>> [SessionCreate] Request json data = [", create_request, "]")
		}

		s := &module.Session{
			UserID:      create_request.UserID,
			WxWebCommon: common.DefaultCommon,
			WxWebXcg:    &conf.XmlConfig{},
			WxApi:       &wxapi.WebwxApi{},
			CreateTime:  time.Now().Unix(),
			LoginStat:   0,
			Loop:        true,
			Quit:        make(chan bool),
		}

		s.UuID, s.QRcode = s.WxApi.WebwxGetUuid(s.WxWebCommon)
        if s.UuID != "" && s.QRcode != "" {
			if glog.V(2) {
				glog.Info(">>> [SessionCreate] UserID ", s.UserID, "'s uuid = ", s.UuID, " && qrcode = ", s.QRcode)
			}

            go s.InitSession(create_request)
            create_response = makeCreateResponse(s.UserID, s.UuID, s.QRcode, 200, "Get uuid from wechat success")
        } else {
            create_response = makeCreateResponse(s.UserID, s.UuID, s.QRcode, -10001, "Get uuid from wechat failed")
        }
	}

	c.JSON(http.StatusOK, create_response)
	return
}

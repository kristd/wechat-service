package handler

import (
    "service/common"
    "service/conf"
    "github.com/gin-gonic/gin"
    "net/http"
    "io"
    "encoding/json"
    "github.com/golang/glog"
    "service/module"
    "time"
    "service/utils"
    "strconv"
    "strings"
)

func makeMassResponse(uid, code int, msg string) *common.Msg_SendAll_Response {
    resp := &common.Msg_SendAll_Response{
        Action: conf.CLIENT_SENDALL,
        UserID: uid,
        Code:   code,
        Msg:    msg,
    }
    return resp
}

func MassMessage(c *gin.Context)  {
    mass_request := &common.Msg_SendAll_Request{}
    mass_response := &common.Msg_SendAll_Response{}

    reqMsgBuf := make([]byte, conf.MAX_BUF_SIZE)
    n, _ := io.ReadFull(c.Request.Body, reqMsgBuf)

    err := json.Unmarshal(reqMsgBuf[:n], mass_request)
    if err != nil {
        glog.Error("[MassMessage] request json data unmarshal err = [", err, "]")
        mass_response = makeMassResponse(mass_request.UserID, -50000, "request json format error")
    } else {
        glog.Info(">>> [MassMessage] Request JSON Data = [", mass_request, "]")
        s, exist := module.SessionTable[mass_request.UserID]
        if !exist {
            glog.Error("[MassMessage] UserID = ", mass_request.UserID, " session not exist")
            mass_response = makeMassResponse(mass_request.UserID, -50001, "Session not exist")
        } else {
            for _, userConf := range s.AutoRepliesConf {
                if userConf.UserType == conf.USER_PERSON {
                    go func() {
                        count := 0

                        for _, contact := range s.ContactMgr.ContactList {
                            if !strings.Contains(contact.UserName, "@") {
                                continue
                            }

                            if userConf.MassText != "" {
                                _, _, err := s.SendText(userConf.MassText, s.Bot.UserName, contact.UserName)
                                if err != nil {
                                    glog.Error("[MassMessage] SendText failed, err = ", err, " NickName = [", contact.NickName, "], userID = [", s.UserID, "]")
                                } else {
                                    glog.Info(">>> [MassMessage] SendText success, NickName = [", contact.NickName, "], userID = [", s.UserID, "]")
                                }
                            }

                            if userConf.MassImage != "" {
                                if len(s.MediaID) == 0  || count % 30 == 0 {
                                    s.MediaID, _ = s.GetMediaID(userConf.MassImage)
                                    glog.Info(">>> [MassMessage] Upload mass image ID = ", s.MediaID)
                                }

                                ret, err := s.SendMassImage(s.MediaID, s.Bot.UserName, contact.UserName)
                                if err != nil || ret != 0 {
                                    glog.Error("[MassMessage] SendMassImage failed, err = ", err, " ret = ", ret, " NickName = [", contact.NickName, "], userID = [", s.UserID, "]")
                                } else {
                                    glog.Info(">>> [MassMessage] SendMassImage success, NickName = [", contact.NickName, "], userID = [", s.UserID, "]")
                                }
                            }

                            n, _ := strconv.Atoi(utils.GetRandomStringFromNum("234", 1))
                            time.Sleep(time.Second * time.Duration(n))

                            count++
                            glog.Info(">>> COUNT <<< ", count)
                        }

                        glog.Info(">>> TOTAL <<< ", count)
                    }()
                }
            }

            mass_response = makeMassResponse(mass_request.UserID, 200, "success")
        }
    }

    c.JSON(http.StatusOK, mass_response)
    return
}
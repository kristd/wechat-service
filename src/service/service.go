package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

var SessionTable map[int]*Session

//func HandleConn(c *gin.Context) {
//	glog.Info(" >>>>> Handle Connection c <<<<<", c.Request)
//
//	reqMsgBuf := make([]byte, 2048)
//
//	n, _ := c.Request.Body.Read(reqMsgBuf)
//	if glog.V(2) {
//		glog.Info(">>>>> Request Body <<<<<", string(reqMsgBuf[:n]))
//	}
//
//	action := GetActionID(reqMsgBuf[:n])
//
//	switch {
//	case action == Client_Action_Create:
//		reqMsg := &Msg_Create_Request{}
//		err := json.Unmarshal(reqMsgBuf[:n], reqMsg)
//		if err != nil {
//			fmt.Println("Client_Action_Create Unmarshal err =", err)
//			return
//		} else {
//			fmt.Println("Client_Action_Create reqMsg =", reqMsg)
//		}
//
//		wxWebXcg := &XmlConfig{}
//		api := &WebwxApi{}
//
//		s := &Session{
//			userID:      reqMsg.userID,
//			wxWebCommon: DefaultCommon,
//			wxWebXcg:    wxWebXcg,
//			wxApi:       api,
//			createTime:  time.Now().Unix(),
//			loginStat:   0,
//		}
//
//		if _, ok := SessionTable[s.userID]; ok {
//			delete(SessionTable, s.userID)
//		}
//		SessionTable[s.userID] = s
//		fmt.Println("SessionTable[uid] =", SessionTable)
//
//		s.autoRepliesConf = make([]AutoReplyConf, len(reqMsg.Config))
//
//		for i := 0; i < len(reqMsg.Config); i++ {
//			s.autoRepliesConf[i].GroupName, _ = reqMsg.Config[i]["group"].(string)
//			sections, succ := reqMsg.Config[i]["keywords"].([]interface{})
//			if succ {
//				s.autoRepliesConf[i].KeyWords = make([]KeyWord, len(sections))
//
//				for j := 0; j < len(sections); j++ {
//					section, ok := sections[j].(map[string]interface{})
//					if ok {
//						key, ok := section["keyword"].(string)
//						if ok {
//							s.autoRepliesConf[i].KeyWords[j].Key = key
//						} else {
//							s.autoRepliesConf[i].KeyWords[j].Key = ""
//							fmt.Println("No Keyword <keyword>")
//						}
//
//						content, ok := section["cotent"].(string)
//						if ok {
//							s.autoRepliesConf[i].KeyWords[j].Text = content
//						} else {
//							s.autoRepliesConf[i].KeyWords[j].Text = ""
//							fmt.Println("No Keyword <cotent>")
//						}
//
//						img, ok := section["Image"].(string)
//						if ok {
//							s.autoRepliesConf[i].KeyWords[j].Image = img
//						} else {
//							s.autoRepliesConf[i].KeyWords[j].Image = ""
//							fmt.Println("No Keyword <Image>")
//						}
//					}
//				}
//			} else {
//				fmt.Println("group <", s.autoRepliesConf[i].GroupName, "> has no keywords")
//			}
//		}
//
//		fmt.Println("s.autoRepliesConf =", s.autoRepliesConf)
//		fmt.Println("SessionTable[uid] =", SessionTable)
//
//		s.uuID, s.qrcode = s.wxApi.WebwxGetUuid(s.wxWebCommon)
//
//		repsMsg := &Msg_Create_Response{
//			Action: Client_Action_Create,
//			userID: s.userID,
//			Uuid:   s.uuID,
//			QrCode: s.qrcode,
//		}
//
//		c.JSON(http.StatusOK, repsMsg)
//		return
//
//		flag := SCAN
//		maxTry := 0
//		redirectUrl, err := s.wxApi.WebwxLogin(s.wxWebCommon, s.uuID, flag)
//		for {
//			if redirectUrl == "201;" {
//
//				fmt.Println("echo 201 =", redirectUrl)
//
//				if flag == SCAN {
//					flag = CONFIRM
//				}
//
//				redirectUrl, err = s.wxApi.WebwxLogin(s.wxWebCommon, s.uuID, flag)
//
//				fmt.Println("echo url ", redirectUrl)
//
//				if err != nil {
//					fmt.Println("WebwxLogin err1 =", err)
//					return
//				} else if strings.Contains(redirectUrl, "http") {
//					s.UpdateLoginStat(LOGIN_SUCC)
//					break
//				}
//			} else if redirectUrl == "408;" {
//
//				fmt.Println("echo 408 =", redirectUrl)
//
//				if flag == CONFIRM {
//					flag = SCAN
//				}
//				redirectUrl, err = s.wxApi.WebwxLogin(s.wxWebCommon, s.uuID, flag)
//				if maxTry > 5 {
//					return
//				} else {
//					maxTry++
//					fmt.Println("maxTry =", maxTry)
//				}
//			} else {
//				fmt.Println("WebwxLogin err2 =", err)
//				s.UpdateLoginStat(LOGIN_FAIL)
//				return
//			}
//		}
//
//		fmt.Println("loginUrl =", redirectUrl)
//
//        s.cookies, err = s.wxApi.WebNewLoginPage(s.wxWebCommon, s.wxWebXcg, redirectUrl)
//		if err != nil {
//			fmt.Println("WebNewLoginPage err =", err)
//		} else {
//			fmt.Println("")
//			fmt.Println(">>>>>cookies <<<<< =", s.cookies)
//		}
//
//		session, err := s.wxApi.WebWxInit(s.wxWebCommon, s.wxWebXcg)
//		if err != nil {
//			fmt.Println("WebWxInit err =", err)
//		} else {
//			fmt.Println("")
//			fmt.Println("")
//			fmt.Println(">>>>>WebWxInit <<<<< ret =", string(session))
//		}
//
//		jc := &JsonConfig{}
//		jc, _ = LoadJsonConfigFromBytes(session)
//
//		sKeyList := &SyncKeyList{}
//
//		sKeyList, err = GetSyncKeyListFromJc(jc)
//		if err != nil {
//			fmt.Println("GetSyncKeyListFromJc err =", err)
//		} else {
//			fmt.Println("")
//			fmt.Println(">>>>>GetSyncKeyListFromJc keylist =", sKeyList)
//		}
//
//		s.bot, err = GetUserInfoFromJc(jc)
//		if err != nil {
//			fmt.Println("GetUserInfoFromJc err =", err)
//		} else {
//			fmt.Println("")
//			fmt.Println(">>>>>USER List<<<<< =", s.bot)
//			fmt.Println(">>>>> User Name <<<<< =", s.bot.UserName)
//		}
//
//		var contacts []byte
//		contacts, err = s.wxApi.WebWxGetContact(s.wxWebCommon, s.wxWebXcg, s.cookies)
//		if err != nil {
//			fmt.Println("WebWxGetContact err =", err)
//		} else {
//			fmt.Println("")
//			fmt.Println(">>>>>Contact List<<<<< =", string(contacts))
//		}
//
//		s.contactMgr, err = CreateContactManagerFromBytes(contacts)
//		if err != nil {
//			fmt.Println(">>>>>CreateContactManagerFromBytes err =", err)
//		}
//
//		s.contactMgr.AddContactFromUser(s.bot)
//
//		/*
//		   添加消息刷新协程
//		*/
//
//	case action == Client_Action_Login:
//		reqMsg := &Msg_Login_Request{}
//		err := json.Unmarshal(reqMsgBuf[:n], reqMsg)
//		if err != nil {
//			fmt.Println("Client_Action_Login err1 =", err)
//		}
//
//		resp := &Msg_Login_Response{}
//		resp.Action = Client_Action_Login
//		resp.userID = reqMsg.userID
//
//		s, ok := SessionTable[reqMsg.userID]
//		if ok {
//			resp.Status = s.loginStat
//		} else {
//			resp.Status = LOGIN_FAIL
//		}
//
//		respBuf, err := json.Marshal(resp)
//		if err != nil {
//			fmt.Println("Client_Action_Login err2 =", err, respBuf)
//		}
//
//		//        w.Write(respBuf)
//	case action == Client_Action_Send:
//		reqMsg := &Msg_Send_Request{}
//		err := json.Unmarshal(reqMsgBuf[:n], reqMsg)
//		if err != nil {
//			fmt.Println("Client_Action_Send err1 =", err)
//		}
//
//		s, ok := SessionTable[reqMsg.userID]
//		if !ok {
//			fmt.Println("Session not exit")
//			return
//		}
//
//		fmt.Println(">>>>> reqMsg <<<<<", reqMsg)
//
//		//ToUsers := s.contactMgr.GetContactByName(reqMsg.Group)
//		toUser := &User{}
//		for _, toUser = range s.contactMgr.contactList {
//			fmt.Println(">>>>> Users Nick Name <<<<<", toUser.NickName)
//			fmt.Println(">>>>> Users User Name <<<<<", toUser.UserName)
//
//			if toUser.NickName == reqMsg.Group {
//				break
//			}
//		}
//
//		var msgID string
//		var localID string
//		var result string
//
//		switch reqMsg.Params.Type {
//		case TEXT_MSG:
//			msgID, localID, err = s.SendText(reqMsg.Params.Content, s.bot.UserName, toUser.UserName)
//			if msgID != "" && localID != "" {
//				fmt.Println("send msg success")
//				result = "success"
//			} else {
//				fmt.Println("SendText err =", err)
//				result = "failed"
//			}
//		case IMG_MSG:
//			fileName, err := LoadImage(reqMsg.Params.Content)
//			if err != nil {
//				fmt.Println("LoadImage err = ", err)
//				result = "failed"
//			} else {
//				retcd, err := s.SendImage(fileName, s.bot.UserName, toUser.UserName)
//				if retcd == 0 {
//					fmt.Println("send image success")
//					result = "success"
//				} else if err != nil {
//					fmt.Println("SendImage err =", err)
//					result = "failed"
//				}
//			}
//		default:
//			result = "failed"
//		}
//
//        fmt.Println(result)
//
//		resp := &Msg_Send_Response{
//			Action: Client_Action_Login,
//			userID: s.userID,
//			Code:   200,
//            Msg:    "",
//		}
//
//		respBuf, err := json.Marshal(resp)
//		if err != nil {
//			fmt.Println("Client_Action_Send err2 =", err, respBuf)
//		}
//
//		//        w.Write(respBuf)
//	case action == Client_Action_Exit:
//		reqMsg := &Msg_Exit_Request{}
//		err := json.Unmarshal(reqMsgBuf[:n], reqMsg)
//		if err != nil {
//			fmt.Println("Client_Action_Exit err1 =", err)
//		}
//
//		s, ok := SessionTable[reqMsg.userID]
//		if ok {
//			s.quit <- true
//			delete(SessionTable, reqMsg.userID)
//		}
//
//		resp := &Msg_Exit_Response{
//			Action: Client_Action_Exit,
//			userID: reqMsg.userID,
//			Code:   200,
//            Msg:    "",
//		}
//
//		respBuf, err := json.Marshal(resp)
//		fmt.Println(respBuf)
//		//        w.Write(respBuf)
//	default:
//		fmt.Println("Action ID Error")
//	}
//}

func main() {
	flag.Parse()
	SessionTable = make(map[int]*Session)

	if glog.V(2) {
		glog.Info("Service Start")
	}

	route := gin.Default()
	route.POST("/api/create", SessionCreate)
	route.POST("/api/login", LoginScan)
	route.POST("/api/send", SendMessage)
	route.POST("/api/exit", Exit)
	route.Run(":8888")
}

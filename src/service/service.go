package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var SessionTable map[int]*Session



type Action_ID struct {
	Action int
}

func GetActionID(data []byte) int {
	actionID := &Action_ID{}

	err := json.Unmarshal(data, actionID)
	if err != nil {
		fmt.Println("GetActionID err =", err)
		return 0
	} else {
		return actionID.Action
	}
}

func LoadImage(url string) (fileName string, err error) {
	path := strings.Split(url, "/")
	if len(path) > 1 {
		fileName = path[len(path)-1]
	}

	fileName = IMG_PATH + fileName
	out, err := os.Create(fileName)

	fmt.Println(">>>>> Image Path <<<< ", fileName)

	defer out.Close()

	resp, err := http.Get(url)
	defer resp.Body.Close()

	pix, err := ioutil.ReadAll(resp.Body)
	_, err = io.Copy(out, bytes.NewReader(pix))

	return fileName, err
}

func WebexMessageSync(s *Session) {
	for {
		select {

		case <-s.quit:
			return
		}
	}
}

// Stop signals the worker to stop listening for work requests.
func Stop(s *Session) {
	go func() {
		s.quit <- true
	}()
}

func PostHandler(c *gin.Context) {
	resp := &Msg_Create_Response{
		Action: 1,
		UserID: 6,
		Uuid:   "test",
		QrCode: "test",
	}

	buf := make([]byte, 2048)
	n, _ := c.Request.Body.Read(buf)
	fmt.Println(string(buf[:n]))

	c.JSON(http.StatusOK, resp)
}

func HandleConn(c *gin.Context) {
	glog.Info(" >>>>> Handle Connection c <<<<<", c.Request)

	reqMsgBuf := make([]byte, 2048)

	n, _ := c.Request.Body.Read(reqMsgBuf)
	if glog.V(2) {
		glog.Info(">>>>> Request Body <<<<<", string(reqMsgBuf[:n]))
	}

	action := GetActionID(reqMsgBuf[:n])

	switch {
	case action == Client_Action_Create:
		reqMsg := &Msg_Create_Request{}
		err := json.Unmarshal(reqMsgBuf[:n], reqMsg)
		if err != nil {
			fmt.Println("Client_Action_Create Unmarshal err =", err)
			return
		} else {
			fmt.Println("Client_Action_Create reqMsg =", reqMsg)
		}

		wxWebXcg := &XmlConfig{}
		api := &WebwxApi{}

		s := &Session{
			UserID:      reqMsg.UserID,
			WxWebCommon: DefaultCommon,
			WxWebXcg:    wxWebXcg,
			wxApi:       api,
			CreateTime:  time.Now().Unix(),
			LoginStat:   0,
		}

		if _, ok := SessionTable[s.UserID]; ok {
			delete(SessionTable, s.UserID)
		}
		SessionTable[s.UserID] = s
		fmt.Println("SessionTable[uid] =", SessionTable)

		s.AutoReplies = make([]AutoReplyConf, len(reqMsg.Config))

		for i := 0; i < len(reqMsg.Config); i++ {
			s.AutoReplies[i].GroupName, _ = reqMsg.Config[i]["group"].(string)
			sections, succ := reqMsg.Config[i]["keywords"].([]interface{})
			if succ {
				s.AutoReplies[i].KeyWords = make([]KeyWord, len(sections))

				for j := 0; j < len(sections); j++ {
					section, ok := sections[j].(map[string]interface{})
					if ok {
						key, ok := section["keyword"].(string)
						if ok {
							s.AutoReplies[i].KeyWords[j].Key = key
						} else {
							s.AutoReplies[i].KeyWords[j].Key = ""
							fmt.Println("No Keyword <keyword>")
						}

						content, ok := section["cotent"].(string)
						if ok {
							s.AutoReplies[i].KeyWords[j].Text = content
						} else {
							s.AutoReplies[i].KeyWords[j].Text = ""
							fmt.Println("No Keyword <cotent>")
						}

						img, ok := section["Image"].(string)
						if ok {
							s.AutoReplies[i].KeyWords[j].Image = img
						} else {
							s.AutoReplies[i].KeyWords[j].Image = ""
							fmt.Println("No Keyword <Image>")
						}
					}
				}
			} else {
				fmt.Println("group <", s.AutoReplies[i].GroupName, "> has no keywords")
			}
		}

		fmt.Println("s.AutoReplies =", s.AutoReplies)
		fmt.Println("SessionTable[uid] =", SessionTable)

		s.UuID, s.Qrcode = s.wxApi.WebwxGetUuid(s.WxWebCommon)

		repsMsg := &Msg_Create_Response{
			Action: Client_Action_Create,
			UserID: s.UserID,
			Uuid:   s.UuID,
			QrCode: s.Qrcode,
		}

		c.JSON(http.StatusOK, repsMsg)
		return

		flag := SCAN
		maxTry := 0
		redirectUrl, err := s.wxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
		for {
			if redirectUrl == "201;" {

				fmt.Println("echo 201 =", redirectUrl)

				if flag == SCAN {
					flag = CONFIRM
				}

				redirectUrl, err = s.wxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)

				fmt.Println("echo url ", redirectUrl)

				if err != nil {
					fmt.Println("WebwxLogin err1 =", err)
					return
				} else if strings.Contains(redirectUrl, "http") {
					s.UpdateLoginStat(LOGIN_SUCC)
					break
				}
			} else if redirectUrl == "408;" {

				fmt.Println("echo 408 =", redirectUrl)

				if flag == CONFIRM {
					flag = SCAN
				}
				redirectUrl, err = s.wxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
				if maxTry > 5 {
					return
				} else {
					maxTry++
					fmt.Println("maxTry =", maxTry)
				}
			} else {
				fmt.Println("WebwxLogin err2 =", err)
				s.UpdateLoginStat(LOGIN_FAIL)
				return
			}
		}

		fmt.Println("loginUrl =", redirectUrl)

        s.Cookies, err = s.wxApi.WebNewLoginPage(s.WxWebCommon, s.WxWebXcg, redirectUrl)
		if err != nil {
			fmt.Println("WebNewLoginPage err =", err)
		} else {
			fmt.Println("")
			fmt.Println(">>>>>Cookies <<<<< =", s.Cookies)
		}

		session, err := s.wxApi.WebWxInit(s.WxWebCommon, s.WxWebXcg)
		if err != nil {
			fmt.Println("WebWxInit err =", err)
		} else {
			fmt.Println("")
			fmt.Println("")
			fmt.Println(">>>>>WebWxInit <<<<< ret =", string(session))
		}

		jc := &JsonConfig{}
		jc, _ = LoadJsonConfigFromBytes(session)

		sKeyList := &SyncKeyList{}

		sKeyList, err = GetSyncKeyListFromJc(jc)
		if err != nil {
			fmt.Println("GetSyncKeyListFromJc err =", err)
		} else {
			fmt.Println("")
			fmt.Println(">>>>>GetSyncKeyListFromJc keylist =", sKeyList)
		}

		s.Bot, err = GetUserInfoFromJc(jc)
		if err != nil {
			fmt.Println("GetUserInfoFromJc err =", err)
		} else {
			fmt.Println("")
			fmt.Println(">>>>>USER List<<<<< =", s.Bot)
			fmt.Println(">>>>> User Name <<<<< =", s.Bot.UserName)
		}

		var contacts []byte
		contacts, err = s.wxApi.WebWxGetContact(s.WxWebCommon, s.WxWebXcg, s.Cookies)
		if err != nil {
			fmt.Println("WebWxGetContact err =", err)
		} else {
			fmt.Println("")
			fmt.Println(">>>>>Contact List<<<<< =", string(contacts))
		}

		s.Cm, err = CreateContactManagerFromBytes(contacts)
		if err != nil {
			fmt.Println(">>>>>CreateContactManagerFromBytes err =", err)
		}

		s.Cm.AddContactFromUser(s.Bot)

		/*
		   添加消息刷新协程
		*/

	case action == Client_Action_Login:
		reqMsg := &Msg_Login_Request{}
		err := json.Unmarshal(reqMsgBuf[:n], reqMsg)
		if err != nil {
			fmt.Println("Client_Action_Login err1 =", err)
		}

		resp := &Msg_Login_Response{}
		resp.Action = Client_Action_Login
		resp.UserID = reqMsg.UserID

		s, ok := SessionTable[reqMsg.UserID]
		if ok {
			resp.Status = s.LoginStat
		} else {
			resp.Status = LOGIN_FAIL
		}

		respBuf, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Client_Action_Login err2 =", err, respBuf)
		}

		//        w.Write(respBuf)
	case action == Client_Action_Send:
		reqMsg := &Msg_Send_Request{}
		err := json.Unmarshal(reqMsgBuf[:n], reqMsg)
		if err != nil {
			fmt.Println("Client_Action_Send err1 =", err)
		}

		s, ok := SessionTable[reqMsg.UserID]
		if !ok {
			fmt.Println("Session not exit")
			return
		}

		fmt.Println(">>>>> reqMsg <<<<<", reqMsg)

		//ToUsers := s.Cm.GetContactByName(reqMsg.Group)
		toUser := &User{}
		for _, toUser = range s.Cm.cl {
			fmt.Println(">>>>> Users Nick Name <<<<<", toUser.NickName)
			fmt.Println(">>>>> Users User Name <<<<<", toUser.UserName)

			if toUser.NickName == reqMsg.Group {
				break
			}
		}

		var msgID string
		var localID string
		var result string

		switch reqMsg.Params.Type {
		case TEXT_MSG:
			msgID, localID, err = s.SendText(reqMsg.Params.Content, s.Bot.UserName, toUser.UserName)
			if msgID != "" && localID != "" {
				fmt.Println("send msg succecss")
				result = "success"
			} else {
				fmt.Println("SendText err =", err)
				result = "failed"
			}
		case IMG_MSG:
			fileName, err := LoadImage(reqMsg.Params.Content)
			if err != nil {
				fmt.Println("LoadImage err = ", err)
				result = "failed"
			} else {
				retcd, err := s.SendImage(fileName, s.Bot.UserName, toUser.UserName)
				if retcd == 0 {
					fmt.Println("send image succecss")
					result = "success"
				} else if err != nil {
					fmt.Println("SendImage err =", err)
					result = "failed"
				}
			}
		default:
			result = "failed"
		}

        fmt.Println(result)

		resp := &Msg_Send_Response{
			Action: Client_Action_Login,
			UserID: s.UserID,
			Code:   200,
            Msg:    "",
		}

		respBuf, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Client_Action_Send err2 =", err, respBuf)
		}

		//        w.Write(respBuf)
	case action == Client_Action_Exit:
		reqMsg := &Msg_Exit_Request{}
		err := json.Unmarshal(reqMsgBuf[:n], reqMsg)
		if err != nil {
			fmt.Println("Client_Action_Exit err1 =", err)
		}

		s, ok := SessionTable[reqMsg.UserID]
		if ok {
			s.quit <- true
			delete(SessionTable, reqMsg.UserID)
		}

		resp := &Msg_Exit_Response{
			Action: Client_Action_Exit,
			UserID: reqMsg.UserID,
			Code:   200,
            Msg:    "",
		}

		respBuf, err := json.Marshal(resp)
		fmt.Println(respBuf)
		//        w.Write(respBuf)
	default:
		fmt.Println("Action ID Error")
	}
}

func main() {
	flag.Parse()
	SessionTable = make(map[int]*Session)

	route := gin.Default()
	route.POST("/api/create", SessionCreate)
	route.POST("/api/login", LoginScan)
	route.POST("/api/send", SendMessage)
	route.POST("/api/exit", Exit)
	route.Run(":8888")
}

package main

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type KeyWord struct {
	Key   string
	Text  string
	Image string
}

type AutoReplyConf struct {
	GroupName string
	KeyWords  []KeyWord
}

type Session struct {
	wxWebCommon *Common
	wxWebXcg    *XmlConfig
	cookies     []*http.Cookie
	synKeyList  *SyncKeyList
	bot         *User
	contactMgr  *ContactManager
	qrcode      string
	uuID        string
	createTime  int64

	//wechat api
	wxApi *WebwxApi

	//user info
	userID          int
	loginStat       int
	autoRepliesConf []AutoReplyConf
	loginMax        int

	redirectUrl string

	//channels
	quit chan bool

	//lock
	statLock sync.RWMutex

	//serve
	stop bool //loop:1/stop:0
}

// SendText: send text msg type 1
func (s *Session) SendText(msg, from, to string) (string, string, error) {
	b, err := s.wxApi.WebWxSendMsg(s.wxWebCommon, s.wxWebXcg, s.cookies, from, to, msg)
	if err != nil {
		return "", "", err
	}
	jc, _ := LoadJsonConfigFromBytes(b)
	ret, _ := jc.GetInt("BaseResponse.Ret")
	if ret != 0 {
		errMsg, _ := jc.GetString("BaseResponse.ErrMsg")
		return "", "", fmt.Errorf("WebWxSendMsg Ret=%d, ErrMsg=%s", ret, errMsg)
	}
	msgID, _ := jc.GetString("MsgID")
	localID, _ := jc.GetString("LocalID")
	return msgID, localID, nil
}

// SendImage: send img, upload then send
func (s *Session) SendImage(path, from, to string) (int, error) {
	ss := strings.Split(path, "/")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		//logs.Error(err)
		return -1, err
	}
	mediaId, err := s.wxApi.WebWxUploadMedia(s.wxWebCommon, s.wxWebXcg, s.cookies, ss[len(ss)-1], b)
	if err != nil {
		//logs.Error(err)
		return -1, fmt.Errorf("Upload image failed")
	}
	ret, err := s.wxApi.WebWxSendMsgImg(s.wxWebCommon, s.wxWebXcg, s.cookies, from, to, mediaId)
	if err != nil || ret != 0 {
		//logs.Error(ret, err)
		return ret, err
	} else {
		return ret, nil
	}
}

func (s *Session) InitSession(request *Msg_Create_Request) {
	if _, ok := SessionTable[s.userID]; ok {
		delete(SessionTable, s.userID)
	}

	SessionTable[s.userID] = s
	fmt.Println("SessionTable =", SessionTable)

	s.autoRepliesConf = make([]AutoReplyConf, len(request.Config))

	for i := 0; i < len(request.Config); i++ {
		s.autoRepliesConf[i].GroupName, _ = request.Config[i]["group"].(string)
		sections, succ := request.Config[i]["keywords"].([]interface{})
		if succ {
			s.autoRepliesConf[i].KeyWords = make([]KeyWord, len(sections))

			for j := 0; j < len(sections); j++ {
				section, ok := sections[j].(map[string]interface{})
				if ok {
					key, ok := section["keyword"].(string)
					if ok {
						s.autoRepliesConf[i].KeyWords[j].Key = key
					} else {
						s.autoRepliesConf[i].KeyWords[j].Key = ""
						fmt.Println("No Keyword <keyword>")
					}

					content, ok := section["cotent"].(string)
					if ok {
						s.autoRepliesConf[i].KeyWords[j].Text = content
					} else {
						s.autoRepliesConf[i].KeyWords[j].Text = ""
						fmt.Println("No Keyword <cotent>")
					}

					img, ok := section["Image"].(string)
					if ok {
						s.autoRepliesConf[i].KeyWords[j].Image = img
					} else {
						s.autoRepliesConf[i].KeyWords[j].Image = ""
						fmt.Println("No Keyword <Image>")
					}
				}
			}
		} else {
			fmt.Println("group <", s.autoRepliesConf[i].GroupName, "> has no keywords")
		}
	}

	fmt.Println("s.autoRepliesConf =", s.autoRepliesConf)
}

func (s *Session) GetLoginStat() int {
	s.statLock.Lock()
	stat := s.loginStat
	s.statLock.Unlock()
	return stat
}

func (s *Session) UpdateLoginStat(stat int) {
	s.statLock.Lock()
	s.loginStat = stat
	s.statLock.Unlock()
}

func (s *Session) StatusPolling(stat chan int) {
	flag := SCAN

	redirectUrl, err := s.wxApi.WebwxLogin(s.wxWebCommon, s.uuID, flag)
	for {
		if redirectUrl == "201;" {
			if flag == SCAN {
				flag = CONFIRM
			}

			fmt.Println("redirectUrl == 201")

			redirectUrl, err = s.wxApi.WebwxLogin(s.wxWebCommon, s.uuID, flag)
			if err != nil {
				fmt.Println(">>> WebwxLogin err1 =", err)
				s.UpdateLoginStat(999)
				stat <- 0
				break
			} else if strings.Contains(redirectUrl, "http") {

				fmt.Println("redirectUrl == ", redirectUrl)

				s.redirectUrl = redirectUrl
				s.UpdateLoginStat(LOGIN_SUCC)
				stat <- 200
				break
			}
		} else if redirectUrl == "408;" {
			s.UpdateLoginStat(LOGIN_FAIL)
			stat <- 2

			fmt.Println("redirectUrl == 408")

			if flag == CONFIRM {
				flag = SCAN
			}
			redirectUrl, err = s.wxApi.WebwxLogin(s.wxWebCommon, s.uuID, flag)
		} else if strings.Contains(redirectUrl, "http") {

			fmt.Println("redirectUrl == ", redirectUrl)

			s.redirectUrl = redirectUrl
			s.UpdateLoginStat(LOGIN_SUCC)
			stat <- 200
			break
		} else {
			fmt.Println(">>> WebwxLogin err2 =", err)
			s.UpdateLoginStat(999)
			stat <- 4
			break
		}
	}
}

func (s *Session) InitUserCookies(redirectUrl string) int {
	var err error

	s.cookies, err = s.wxApi.WebNewLoginPage(s.wxWebCommon, s.wxWebXcg, redirectUrl)
	if err != nil {
		fmt.Println("WebNewLoginPage err =", err)
		return -1
	} else {
		fmt.Println("")
		fmt.Println(">>>>>cookies <<<<< =", s.cookies)
	}

	session, err := s.wxApi.WebWxInit(s.wxWebCommon, s.wxWebXcg)
	if err != nil {
		fmt.Println("WebWxInit err =", err)
		return -2
	} else {
		fmt.Println("")
		fmt.Println("")
		fmt.Println(">>>>>WebWxInit <<<<< ret =", string(session))
	}

	jc := &JsonConfig{}
	jc, _ = LoadJsonConfigFromBytes(session)

	s.synKeyList, err = GetSyncKeyListFromJc(jc)
	if err != nil {
		fmt.Println("GetSyncKeyListFromJc err =", err)
		return -3
	} else {
		fmt.Println("")
		fmt.Println(">>>>>GetSyncKeyListFromJc keylist =", s.synKeyList)
	}

	s.bot, err = GetUserInfoFromJc(jc)
	if err != nil {
		fmt.Println("GetUserInfoFromJc err =", err)
		return -4
	} else {
		fmt.Println("")
		fmt.Println(">>>>>USER List<<<<< =", s.bot)
		fmt.Println(">>>>> User Name <<<<< =", s.bot.UserName)
	}

	var contacts []byte
	contacts, err = s.wxApi.WebWxGetContact(s.wxWebCommon, s.wxWebXcg, s.cookies)
	if err != nil {
		fmt.Println("WebWxGetContact err =", err)
		return -5
	} else {
		fmt.Println("")
		fmt.Println(">>>>>Contact List<<<<< =", string(contacts))
	}

	s.contactMgr, err = CreateContactManagerFromBytes(contacts)
	if err != nil {
		fmt.Println(">>>>>CreateContactManagerFromBytes err =", err)
		return -6
	}

	s.contactMgr.AddContactFromUser(s.bot)
	return 0
}

func (s *Session) InitAndServe() {
	ret := s.InitUserCookies(s.redirectUrl)

	fmt.Println("After s.InitUserCookies(s.redirectUrl) ")

	if ret == 0 {
		go s.Serve()
	} else {
		glog.Error("init cookies failed =", ret)
	}
}

func (s *Session) Serve() {

	fmt.Println("")
	fmt.Println(">>> Session Serving <<<")

	for !s.stop {
		ret, selector, err := s.wxApi.SyncCheck(s.wxWebCommon, s.wxWebXcg, s.cookies, s.wxWebCommon.SyncSrv, s.synKeyList)

		fmt.Println("ret =", ret, "||select =", selector, "||err =", err)

		if err != nil {
			glog.Info(">>> SyncCheck err =", err)
			continue
		}
		if ret == 0 {
			// check success
			if selector == 2 {
				// new message

				fmt.Println(">>> Before s.wxApi.WebWxSync")

				msg, err := s.wxApi.WebWxSync(s.wxWebCommon, s.wxWebXcg, s.cookies, s.synKeyList)

				fmt.Println(">>> After s.wxApi.WebWxSync")
				fmt.Println(">>> msg, err <<<", msg, "||", err)

				if err != nil {
					fmt.Println("WebWxSync err", err)
				} else {
					//fmt.Println(">>> Receive message <<< =", string(msg))

					jc, err := LoadJsonConfigFromBytes(msg)
					if err != nil {
						fmt.Println(">>> LoadJsonConfigFromBytes err =", err)
					}

					msgCount, _ := jc.GetInt("AddMsgCount")
					if msgCount < 1 {
						fmt.Println(">>> MsgCount == 0")
						continue
					}
					msgis, _ := jc.GetInterfaceSlice("AddMsgList")
					for _, v := range msgis {
						rmsg := s.Analize(v.(map[string]interface{}))

						fmt.Println(rmsg.FromUserName, "<<< >>> rmsg.Content = <<< ", rmsg.Content)

						go s.ReplyMsg(rmsg.FromUserName, rmsg.Content)
					}
				}
			} else if selector != 0 && selector != 7 {
				fmt.Println("session down, sel %d", selector)
			}
		} else if ret == 1101 {
			fmt.Println("error code =", ret)
		} else if ret == 1205 {
			fmt.Println("api blocked, ret:%d", 1205)
		} else {
			fmt.Println("unhandled exception ret %d", ret)
		}
	}

	fmt.Println("")
	fmt.Println(">>> Serve Stop <<<")
	s.quit <- true
    fmt.Println(">>> Write Chan")
}

func (s *Session) Analize(msg map[string]interface{}) *ReceivedMessage {
	rmsg := &ReceivedMessage{
		MsgId:         msg["MsgId"].(string),
		OriginContent: msg["Content"].(string),
		FromUserName:  msg["FromUserName"].(string),
		ToUserName:    msg["ToUserName"].(string),
		MsgType:       int(msg["MsgType"].(float64)),
	}

	if rmsg.MsgType == MSG_FV {
		riif := msg["RecommendInfo"].(map[string]interface{})
		rmsg.RecommendInfo = &RecommendInfo{
			Ticket:   riif["Ticket"].(string),
			UserName: riif["UserName"].(string),
			NickName: riif["NickName"].(string),
			Content:  riif["Content"].(string),
			Sex:      int(riif["Sex"].(float64)),
		}
	}

	if strings.Contains(rmsg.FromUserName, "@@") ||
		strings.Contains(rmsg.ToUserName, "@@") {
		rmsg.IsGroup = true
		// group message
		ss := strings.Split(rmsg.OriginContent, ":<br/>")
		if len(ss) > 1 {
			rmsg.Who = ss[0]
			rmsg.Content = ss[1]
		} else {
			rmsg.Who = s.bot.UserName
			rmsg.Content = rmsg.OriginContent
		}
	} else {
		// no group message
		rmsg.Who = rmsg.FromUserName
		rmsg.Content = rmsg.OriginContent
	}

	if rmsg.MsgType == MSG_TEXT &&
		len(rmsg.Content) > 1 &&
		strings.HasPrefix(rmsg.Content, "@") {
		// @someone
		ss := strings.Split(rmsg.Content, "\u2005")
		if len(ss) == 2 {
			rmsg.At = ss[0] + "\u2005"
			rmsg.Content = ss[1]
		}
	}
	return rmsg
}

func (s *Session) ReplyMsg(group, msg string) {
	toUser := &User{}
	match := false

	for _, toUser = range s.contactMgr.contactList {

		fmt.Println(">>> User Name =", toUser.UserName, "||", toUser.NickName)

		if toUser.UserName == group {

			fmt.Println(">>> Match")

			match = true
			break
		}
	}

	if match {
		for _, groupConf := range s.autoRepliesConf {

			fmt.Println("s.autoRepliesConf =", groupConf.GroupName)

			if groupConf.GroupName == toUser.NickName {

				fmt.Println("A")

				for _, keyword := range groupConf.KeyWords {

					fmt.Println(">>> msg =", msg, " || range groupConf.KeyWords =", keyword.Key)

					if strings.Contains(msg, keyword.Key) {

						fmt.Println("B")

						if keyword.Text != "" {
							msgID, localID, err := s.SendText(keyword.Text, s.bot.UserName, toUser.UserName)
							if err != nil {
								fmt.Println("text err =", err)
							} else {
								fmt.Println("msgID & localID =", msgID, " || ", localID)
							}
						}

						if keyword.Image != "" {
							ret, err := s.SendImage("/Users/kristd/Documents/sublime/image/logo.png", s.bot.UserName, toUser.UserName)
							if err != nil {
								fmt.Println("image err =", err)
							} else {
								fmt.Println("retcd =", ret)
							}
						}
					}
				}
			}
		}
	}
}

func (s *Session) Stop() {
	s.stop = true
	fmt.Println(">>> loop set 0 <<<")
}

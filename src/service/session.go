package main

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
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
	WxWebCommon *Common
	WxWebXcg    *XmlConfig
	Cookies     []*http.Cookie
	SynKeyList  *SyncKeyList
	Bot         *User
	Cm          *ContactManager
	Qrcode      string
	UuID        string
	CreateTime  int64

	//wechat api
	wxApi *WebwxApi

	//user info
	UserID int
	//UserPushMsgList []*UserPushMsg
	LoginStat       int
	AutoRepliesConf []AutoReplyConf
	loginMax        int

	redirectUrl string

	//channels
	quit chan bool
	push chan string
	pull chan string

	//lock
	statLock sync.RWMutex
}

// SendText: send text msg type 1
func (s *Session) SendText(msg, from, to string) (string, string, error) {
	b, err := s.wxApi.WebWxSendMsg(s.WxWebCommon, s.WxWebXcg, s.Cookies, from, to, msg)
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
	mediaId, err := s.wxApi.WebWxUploadMedia(s.WxWebCommon, s.WxWebXcg, s.Cookies, ss[len(ss)-1], b)
	if err != nil {
		//logs.Error(err)
		return -1, fmt.Errorf("Upload image failed")
	}
	ret, err := s.wxApi.WebWxSendMsgImg(s.WxWebCommon, s.WxWebXcg, s.Cookies, from, to, mediaId)
	if err != nil || ret != 0 {
		//logs.Error(ret, err)
		return ret, err
	} else {
		return ret, nil
	}
}

func (s *Session) InitSession(request *Msg_Create_Request) {
	if _, ok := SessionTable[s.UserID]; ok {
		delete(SessionTable, s.UserID)
	}

	SessionTable[s.UserID] = s
	fmt.Println("SessionTable =", SessionTable)

	s.AutoRepliesConf = make([]AutoReplyConf, len(request.Config))

	for i := 0; i < len(request.Config); i++ {
		s.AutoRepliesConf[i].GroupName, _ = request.Config[i]["group"].(string)
		sections, succ := request.Config[i]["keywords"].([]interface{})
		if succ {
			s.AutoRepliesConf[i].KeyWords = make([]KeyWord, len(sections))

			for j := 0; j < len(sections); j++ {
				section, ok := sections[j].(map[string]interface{})
				if ok {
					key, ok := section["keyword"].(string)
					if ok {
						s.AutoRepliesConf[i].KeyWords[j].Key = key
					} else {
						s.AutoRepliesConf[i].KeyWords[j].Key = ""
						fmt.Println("No Keyword <keyword>")
					}

					content, ok := section["cotent"].(string)
					if ok {
						s.AutoRepliesConf[i].KeyWords[j].Text = content
					} else {
						s.AutoRepliesConf[i].KeyWords[j].Text = ""
						fmt.Println("No Keyword <cotent>")
					}

					img, ok := section["Image"].(string)
					if ok {
						s.AutoRepliesConf[i].KeyWords[j].Image = img
					} else {
						s.AutoRepliesConf[i].KeyWords[j].Image = ""
						fmt.Println("No Keyword <Image>")
					}
				}
			}
		} else {
			fmt.Println("group <", s.AutoRepliesConf[i].GroupName, "> has no keywords")
		}
	}

	fmt.Println("s.AutoRepliesConf =", s.AutoRepliesConf)
}

func (s *Session) GetLoginStat() int {
	s.statLock.Lock()
	stat := s.LoginStat
	s.statLock.Unlock()
	return stat
}

func (s *Session) UpdateLoginStat(stat int) {
	s.statLock.Lock()
	s.LoginStat = stat
	s.statLock.Unlock()
}

func (s *Session) StatusPolling(stat chan int) {
	flag := SCAN

	redirectUrl, err := s.wxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
	for {
		if redirectUrl == "201;" {
			if flag == SCAN {
				flag = CONFIRM
			}

			fmt.Println("redirectUrl == 201")

			redirectUrl, err = s.wxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
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
			redirectUrl, err = s.wxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
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

	s.Cookies, err = s.wxApi.WebNewLoginPage(s.WxWebCommon, s.WxWebXcg, redirectUrl)
	if err != nil {
		fmt.Println("WebNewLoginPage err =", err)
		return -1
	} else {
		fmt.Println("")
		fmt.Println(">>>>>Cookies <<<<< =", s.Cookies)
	}

	session, err := s.wxApi.WebWxInit(s.WxWebCommon, s.WxWebXcg)
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

	s.SynKeyList, err = GetSyncKeyListFromJc(jc)
	if err != nil {
		fmt.Println("GetSyncKeyListFromJc err =", err)
		return -3
	} else {
		fmt.Println("")
		fmt.Println(">>>>>GetSyncKeyListFromJc keylist =", s.SynKeyList)
	}

	s.Bot, err = GetUserInfoFromJc(jc)
	if err != nil {
		fmt.Println("GetUserInfoFromJc err =", err)
		return -4
	} else {
		fmt.Println("")
		fmt.Println(">>>>>USER List<<<<< =", s.Bot)
		fmt.Println(">>>>> User Name <<<<< =", s.Bot.UserName)
	}

	var contacts []byte
	contacts, err = s.wxApi.WebWxGetContact(s.WxWebCommon, s.WxWebXcg, s.Cookies)
	if err != nil {
		fmt.Println("WebWxGetContact err =", err)
		return -5
	} else {
		fmt.Println("")
		fmt.Println(">>>>>Contact List<<<<< =", string(contacts))
	}

	s.Cm, err = CreateContactManagerFromBytes(contacts)
	if err != nil {
		fmt.Println(">>>>>CreateContactManagerFromBytes err =", err)
		return -6
	}

	s.Cm.AddContactFromUser(s.Bot)
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
	fmt.Println("Session Serving")

	for {
        select {
        case <- s.quit:
            return
        }

		time.Sleep(1 * time.Second)

		ret, selector, err := s.wxApi.SyncCheck(s.WxWebCommon, s.WxWebXcg, s.Cookies, s.WxWebCommon.SyncSrv, s.SynKeyList)
		if err != nil {
			glog.Error(err)
			continue
		}
		if ret == 0 {
			// check success
			if selector == 2 {
				// new message
				msg, err := s.wxApi.WebWxSync(s.WxWebCommon, s.WxWebXcg, s.Cookies, s.SynKeyList)
				if err != nil {
					glog.Error(err)
				} else {
					fmt.Println(">>> Receive message <<< =", string(msg))
                    // analize message
                    jc, _ := LoadJsonConfigFromBytes(msg)
                    msgCount, _ := jc.GetInt("AddMsgCount")
                    if msgCount < 1 {
                        // no msg
                        return
                    }
                    msgis, _ := jc.GetInterfaceSlice("AddMsgList")
                    for _, v := range msgis {
                        rmsg := s.Analize(v.(map[string]interface{}))
                        //err, handles := s.HandlerRegister.Get(rmsg.MsgType)
                        if err != nil {
                            glog.Error(err)
                            continue
                        } else {
                            fmt.Println(rmsg.FromUserName, "<<< >>> rmsg.Content = <<< ", rmsg.Content)
                        }
                        go s.ReplyMsg(rmsg.FromUserName, rmsg.Content)
                    }
				}
			} else if selector != 0 && selector != 7 {
				glog.Error("session down, sel %d", selector)
				//break loop1
			}
		} else if ret == 1101 {
			//errChan <- nil
			//break loop1
		} else if ret == 1205 {
			glog.Error("api blocked, ret:%d", 1205)
			//break loop1
		} else {
			glog.Error("unhandled exception ret %d", ret)
			//break loop1
		}
	}
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
            rmsg.Who = s.Bot.UserName
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

    for _, toUser = range s.Cm.cl {
        if toUser.UserName == group {
            match = true
            break
        }
    }

   if match {
        for _, groupConf := range s.AutoRepliesConf {
            if groupConf.GroupName == toUser.NickName {
                for _, keyword := range groupConf.KeyWords {
                    if strings.Contains(msg, keyword.Key) {
                        if keyword.Text != "" {
                            msgID, localID, err := s.SendText(keyword.Text, s.Bot.UserName, toUser.UserName)
                            if err != nil {
                                fmt.Println("text err =", err)
                            } else {
                                fmt.Println("msgID & localID =", msgID, " || ", localID)
                            }
                        }

                        if keyword.Image != "" {
                            ret, err := s.SendImage("/Users/kristd/Documents/sublime/image/logo.png", s.Bot.UserName, toUser.UserName)
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
    s.quit <- true
}
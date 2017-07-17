package module

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"service/wxapi"
    "service/conf"
    "service/common"
)

type keyWord struct {
	key   string
	text  string
	image string
}

type autoReplyConf struct {
	groupName string
	keyWords  []keyWord
}

type Session struct {
	WxWebCommon *common.Common
	WxWebXcg    *conf.XmlConfig
	cookies     []*http.Cookie
	synKeyList  *common.SyncKeyList
	Bot         *common.User
	ContactMgr  *common.ContactManager
	QRcode      string
	UuID        string
	CreateTime  int64

	//wechat api
	WxApi *wxapi.WebwxApi

	//user info
	UserID          int
	LoginStat       int
	autoRepliesConf []autoReplyConf

	redirectUrl string

	//channels
	Quit chan bool

	//lock
	statLock sync.RWMutex

	//serve
	Loop bool
}

var SessionTable map[int]*Session

// SendText: send text msg type 1
func (s *Session) SendText(msg, from, to string) (string, string, error) {
	b, err := s.WxApi.WebWxSendMsg(s.WxWebCommon, s.WxWebXcg, s.cookies, from, to, msg)
	if err != nil {
		return "", "", err
	}
	jc, _ := conf.LoadJsonConfigFromBytes(b)
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
	mediaId, err := s.WxApi.WebWxUploadMedia(s.WxWebCommon, s.WxWebXcg, s.cookies, ss[len(ss)-1], b)
	if err != nil {
		//logs.Error(err)
		return -1, fmt.Errorf("Upload image failed")
	}
	ret, err := s.WxApi.WebWxSendMsgImg(s.WxWebCommon, s.WxWebXcg, s.cookies, from, to, mediaId)
	if err != nil || ret != 0 {
		//logs.Error(ret, err)
		return ret, err
	} else {
		return ret, nil
	}
}

func (s *Session) InitSession(request *common.Msg_Create_Request) {
	if _, exist := SessionTable[s.UserID]; exist {
		if glog.V(2) {
            glog.Info(">>> [InitSession] Delete UserID ", s.UserID, "'s session")
        }
		delete(SessionTable, s.UserID)
	}

	SessionTable[s.UserID] = s
	s.autoRepliesConf = make([]autoReplyConf, len(request.Config))

	for i := 0; i < len(request.Config); i++ {
		s.autoRepliesConf[i].groupName, _ = request.Config[i]["group"].(string)
		sections, exist := request.Config[i]["keywords"].([]interface{})
		if exist {
			s.autoRepliesConf[i].keyWords = make([]keyWord, len(sections))

			for j := 0; j < len(sections); j++ {
				section, exist := sections[j].(map[string]interface{})
				if exist {
					key, ok := section["keyword"].(string)
					if ok {
						s.autoRepliesConf[i].keyWords[j].key = key
					} else {
						s.autoRepliesConf[i].keyWords[j].key = ""
					}

					content, exist := section["text"].(string)
					if exist {
						s.autoRepliesConf[i].keyWords[j].text = content
					} else {
						s.autoRepliesConf[i].keyWords[j].text = ""
					}

					img, exist := section["image"].(string)
					if exist {
						s.autoRepliesConf[i].keyWords[j].image = img
					} else {
						s.autoRepliesConf[i].keyWords[j].image = ""
					}
				}
			}

            if glog.V(2) {
                glog.Info(">>> [InitSession] Group [", s.autoRepliesConf[i].groupName, "] keyword configs = [", s.autoRepliesConf, "]")
            }
		} else {
            if glog.V(2) {
			    glog.Info(">>> [InitSession] Group [", s.autoRepliesConf[i].groupName, "] has no keywords")
            }
		}
	}
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
	flag := conf.SCAN

	redirectUrl, err := s.WxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
	for {
		if redirectUrl == "201;" {
			if flag == conf.SCAN {
				flag = conf.CONFIRM
			}

			fmt.Println("redirectUrl == 201")

			redirectUrl, err = s.WxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
			if err != nil {
				fmt.Println(">>> WebwxLogin err1 =", err)
				s.UpdateLoginStat(999)
				stat <- 0
				break
			} else if strings.Contains(redirectUrl, "http") {

				fmt.Println("redirectUrl == ", redirectUrl)

				s.redirectUrl = redirectUrl
				s.UpdateLoginStat(conf.LOGIN_SUCC)
				stat <- 200
				break
			}
		} else if redirectUrl == "408;" {
			s.UpdateLoginStat(conf.LOGIN_FAIL)
			stat <- 2

			fmt.Println("redirectUrl == 408")

			if flag == conf.CONFIRM {
				flag = conf.SCAN
			}
			redirectUrl, err = s.WxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
		} else if strings.Contains(redirectUrl, "http") {

			fmt.Println("redirectUrl == ", redirectUrl)

			s.redirectUrl = redirectUrl
			s.UpdateLoginStat(conf.LOGIN_SUCC)
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

	s.cookies, err = s.WxApi.WebNewLoginPage(s.WxWebCommon, s.WxWebXcg, redirectUrl)
	if err != nil {
		fmt.Println("WebNewLoginPage err =", err)
		return -1
	} else {
		fmt.Println("")
		fmt.Println(">>>>>cookies <<<<< =", s.cookies)
	}

	session, err := s.WxApi.WebWxInit(s.WxWebCommon, s.WxWebXcg)
	if err != nil {
		fmt.Println("WebWxInit err =", err)
		return -2
	} else {
		fmt.Println("")
		fmt.Println("")
		fmt.Println(">>>>>WebWxInit <<<<< ret =", string(session))
	}

	jc := &conf.JsonConfig{}
	jc, _ = conf.LoadJsonConfigFromBytes(session)

	s.synKeyList, err = common.GetSyncKeyListFromJc(jc)
	if err != nil {
		fmt.Println("GetSyncKeyListFromJc err =", err)
		return -3
	} else {
		fmt.Println("")
		fmt.Println(">>>>>GetSyncKeyListFromJc keylist =", s.synKeyList)
	}

	s.Bot, err = common.GetUserInfoFromJc(jc)
	if err != nil {
		fmt.Println("GetUserInfoFromJc err =", err)
		return -4
	} else {
		fmt.Println("")
		fmt.Println(">>>>>USER List<<<<< =", s.Bot)
		fmt.Println(">>>>> User Name <<<<< =", s.Bot.UserName)
	}

	var contacts []byte
	contacts, err = s.WxApi.WebWxGetContact(s.WxWebCommon, s.WxWebXcg, s.cookies)
	if err != nil {
		fmt.Println("WebWxGetContact err =", err)
		return -5
	} else {
		fmt.Println("")
		fmt.Println(">>>>>Contact List<<<<< =", string(contacts))
	}

	s.ContactMgr, err = common.CreateContactManagerFromBytes(contacts)
	if err != nil {
		fmt.Println(">>>>>CreateContactManagerFromBytes err =", err)
		return -6
	}

	s.ContactMgr.AddContactFromUser(s.Bot)
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

	for s.Loop {
		ret, selector, err := s.WxApi.SyncCheck(s.WxWebCommon, s.WxWebXcg, s.cookies, s.WxWebCommon.SyncSrv, s.synKeyList)

		fmt.Println("ret =", ret, "||select =", selector, "||err =", err)

		if err != nil {
			glog.Info(">>> SyncCheck err =", err)
			continue
		}
		if ret == 0 {
			// check success
			if selector == 2 {
				// new message

				fmt.Println(">>> Before s.WxApi.WebWxSync")

				msg, err := s.WxApi.WebWxSync(s.WxWebCommon, s.WxWebXcg, s.cookies, s.synKeyList)

				fmt.Println(">>> After s.WxApi.WebWxSync")

				if err != nil {
					fmt.Println("WebWxSync err", err)
				} else {
					jc, err := conf.LoadJsonConfigFromBytes(msg)
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

    fmt.Println(">>> Serving Stop <<<")
	s.Quit <- true
}

func (s *Session) Analize(msg map[string]interface{}) *common.ReceivedMessage {
	rmsg := &common.ReceivedMessage{
		MsgId:         msg["MsgId"].(string),
		OriginContent: msg["Content"].(string),
		FromUserName:  msg["FromUserName"].(string),
		ToUserName:    msg["ToUserName"].(string),
		MsgType:       int(msg["MsgType"].(float64)),
	}

	if rmsg.MsgType == conf.MSG_FV {
		riif := msg["RecommendInfo"].(map[string]interface{})
		rmsg.RecommendInfo = &common.RecommendInfo{
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

	if rmsg.MsgType == conf.MSG_TEXT &&
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
	toUser := &common.User{}
	match := false

	for _, toUser = range s.ContactMgr.ContactList {

		fmt.Println(">>> User Name =", toUser.UserName, "||", toUser.NickName)

		if toUser.UserName == group {
			match = true
			break
		}
	}

	if match {
		for _, groupConf := range s.autoRepliesConf {

			fmt.Println("s.autoRepliesConf =", groupConf.groupName)

			if groupConf.groupName == toUser.NickName {
				for _, keyword := range groupConf.keyWords {

					fmt.Println(">>> msg =", msg, " || range groupConf.keyWords =", keyword.key)

					if strings.Contains(msg, keyword.key) {
						if keyword.text != "" {
							msgID, localID, err := s.SendText(keyword.text, s.Bot.UserName, toUser.UserName)
							if err != nil {
								fmt.Println("text err =", err)
							} else {
								fmt.Println("msgID & localID =", msgID, " || ", localID)
							}
						}

						if keyword.image != "" {
							ret, err := s.SendImage("./image/logo.png", s.Bot.UserName, toUser.UserName)
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
	s.Loop = false
}

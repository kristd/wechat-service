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
	"net/url"
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

	RedirectUrl string

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
		return -1, err
	}
	mediaId, err := s.WxApi.WebWxUploadMedia(s.WxWebCommon, s.WxWebXcg, s.cookies, ss[len(ss)-1], b)
	if err != nil {
		return -1, fmt.Errorf("Upload image failed")
	}
	ret, err := s.WxApi.WebWxSendMsgImg(s.WxWebCommon, s.WxWebXcg, s.cookies, from, to, mediaId)
	if err != nil || ret != 0 {
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

func (s *Session) LoginPolling() int {
    tryCount := 0
	flag := conf.SCAN

	for {
		redirectUrl, err := s.WxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
		if err != nil {
			if glog.V(2) {
				glog.Error("[LoginPolling] sec1 WebwxLogin failed, uuid = ", s.UuID, " err = [", err, "]")
			}
		} else if redirectUrl == "201;" {
			if flag == conf.SCAN {
				flag = conf.CONFIRM
			}

			if glog.V(2) {
				glog.Info("[LoginPolling] sec1 WebwxLogin uuid = ", s.UuID, " redirectUrl = ", redirectUrl)
			}

			redirectUrl, err = s.WxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
			if err != nil {
                if glog.V(2) {
                    glog.Error("[LoginPolling] sec2 WebwxLogin failed, uuid = ", s.UuID, ", err = [", err, "]")
                }

				s.UpdateLoginStat(-200)

				return -200
			} else if strings.Contains(redirectUrl, "http") {
				s.RedirectUrl = redirectUrl
				s.UpdateLoginStat(conf.LOGIN_SUCC)

                if glog.V(2) {
                    glog.Info("[LoginPolling] sec2 WebwxLogin success, uuid = ", s.UuID)
                }

				return 200
			}
		} else if redirectUrl == "408;" {
			s.UpdateLoginStat(conf.LOGIN_FAIL)
			if flag == conf.CONFIRM {
				flag = conf.SCAN
			}

			if glog.V(2) {
				glog.Info("[LoginPolling] sec3 WebwxLogin uuid = ", s.UuID, " redirectUrl = ", redirectUrl)
			}

			tryCount++
			if tryCount >= conf.MAXTRY {
				s.UpdateLoginStat(-998)
				return -998
			} else {
				if glog.V(2) {
					glog.Info("[LoginPolling] sec3 WebwxLogin uuid = ", s.UuID, " retry =", tryCount)
				}
			}
		} else if strings.Contains(redirectUrl, "http") {
			s.RedirectUrl = redirectUrl
			s.UpdateLoginStat(conf.LOGIN_SUCC)

			if glog.V(2) {
				glog.Info("[LoginPolling] sec3 WebwxLogin success, uuid = ", s.UuID)
			}

			return 200
		} else {
			s.UpdateLoginStat(-999)

			if glog.V(2) {
				glog.Error("[LoginPolling] sec4 WebwxLogin failed, uuid = ", s.UuID, ", redirectUrl = ", redirectUrl)
			}

			return -999
		}
	}
}

func (s *Session) AnalizeVersion(uri string) {
	u, _ := url.Parse(uri)

	// version may change
	s.WxWebCommon.CgiDomain = u.Scheme + "://" + u.Host
	s.WxWebCommon.CgiUrl = s.WxWebCommon.CgiDomain + "/cgi-bin/mmwebwx-bin"

	if strings.Contains(u.Host, "wx2") {
		// new version
		s.WxWebCommon.SyncSrv = "webpush.wx2.qq.com"
	} else {
		// old version
		s.WxWebCommon.SyncSrv = "webpush.wx.qq.com"
	}
}

func (s *Session) InitUserCookies(redirectUrl string) int {
	var err error

	s.cookies, err = s.WxApi.WebNewLoginPage(s.WxWebCommon, s.WxWebXcg, redirectUrl)
	if err != nil {
		if glog.V(2) {
			glog.Error("[InitUserCookies] WebNewLoginPage err = ", err)
		}
		return -1
	}

	session, err := s.WxApi.WebWxInit(s.WxWebCommon, s.WxWebXcg)
	if err != nil {
		if glog.V(2) {
			glog.Error("[InitUserCookies] WebWxInit err =", err)
		}
		return -2
	}


	fmt.Println("")
	fmt.Println("")
	fmt.Println(">>> SESSION DATA =[", string(session), "]")
	fmt.Println("")
	fmt.Println("")


	jc := &conf.JsonConfig{}
	jc, _ = conf.LoadJsonConfigFromBytes(session)

	s.synKeyList, err = common.GetSyncKeyListFromJc(jc)
	if err != nil {
		if glog.V(2) {
			glog.Error("[InitUserCookies] GetSyncKeyListFromJc err =", err)
		}
		return -3
	}

	s.Bot, err = common.GetUserInfoFromJc(jc)
	if err != nil {
		if glog.V(2) {
			glog.Error("[InitUserCookies] GetUserInfoFromJc err =", err)
		}
		return -4
	}

	var contacts []byte
	contacts, err = s.WxApi.WebWxGetContact(s.WxWebCommon, s.WxWebXcg, s.cookies)
	if err != nil {
		if glog.V(2) {
			glog.Error("[InitUserCookies] WebWxGetContact err =", err)
		}
		return -5
	}

	s.ContactMgr, err = common.CreateContactManagerFromBytes(contacts)
	if err != nil {
		if glog.V(2) {
			glog.Error("[InitUserCookies] CreateContactManagerFromBytes err =", err)
		}
		return -6
	}

	s.ContactMgr.AddContactFromUser(s.Bot)
	return 0
}

func (s *Session) Serve() {
	fmt.Println(">>> [Serve] Serving start userID = ", s.UserID)

	for s.Loop {
        //Will be blocked here until wechat return response
		ret, selector, err := s.WxApi.SyncCheck(s.WxWebCommon, s.WxWebXcg, s.cookies, s.WxWebCommon.SyncSrv, s.synKeyList)
		if err != nil {
			glog.Info(">>> SyncCheck err =", err)
			continue
		}
		if ret == 0 {
			// check success
			if selector == 2 {
				// new message
				msg, err := s.WxApi.WebWxSync(s.WxWebCommon, s.WxWebXcg, s.cookies, s.synKeyList)
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
			break
		} else if ret == 1205 {
			fmt.Println("api blocked, ret:%d", 1205)
			break
		} else {
			fmt.Println("unhandled exception ret %d", ret)
			break
		}
	}

    fmt.Println(">>> [Serve] Serving stop userID = ", s.UserID)
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
		if toUser.UserName == group {
			match = true
			break
		}
	}

	if match {
		for _, groupConf := range s.autoRepliesConf {
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

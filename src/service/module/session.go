package module

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"service/common"
	"service/conf"
	"service/utils"
	"service/wxapi"
	"strings"
	"sync"
)

type KeyWord struct {
	Key   string
	Text  string
	Image string
}

type AutoReplyConf struct {
	NickName string
	UserType int
	WlmText  string
	WlmImage string
	KeyWords []KeyWord
}

type Session struct {
	WxWebCommon *common.Common
	WxWebXcg    *conf.XmlConfig
	Cookies     []*http.Cookie
	SynKeyList  *common.SyncKeyList
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
	AutoRepliesConf []AutoReplyConf

	RedirectUrl string

	//channels
	Quit chan bool

	//lock
	statLock  sync.RWMutex
	serveLock sync.RWMutex

	//serve
	Loop bool
}

var (
	SessionTable map[int]*Session
)

// SendText: send text msg type 1
func (s *Session) SendText(msg, from, to string) (string, string, error) {
	b, err := s.WxApi.WebWxSendMsg(s.WxWebCommon, s.WxWebXcg, s.Cookies, from, to, msg)
	if err != nil {
		return "", "", err
	}
	jc, _ := conf.LoadJsonConfigFromBytes(b)
	ret, _ := jc.GetInt("BaseResponse.Ret")
	if ret != 0 {
		errMsg, _ := jc.GetString("BaseResponse.ErrMsg")
		glog.Error("[SendText] WebWxSendMsg Ret = ", ret, " ErrMsg = ", errMsg)
		return "", "", fmt.Errorf("[SendText] WebWxSendMsg Ret=%d, ErrMsg=%s", ret, errMsg)
	}
	msgID, _ := jc.GetString("MsgID")
	localID, _ := jc.GetString("LocalID")
	return msgID, localID, nil
}

// SendImage: send img, upload then send
func (s *Session) SendImage(path, from, to string) (int, error) {
	fileName, err := utils.LoadImage(path)
	if err != nil {
		glog.Error("[SendImage] Download image failed, err = ", err)
		return -1, fmt.Errorf("[SendImage] Download image failed, err = ", err)
	}

	ss := strings.Split(fileName, "/")
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return -2, err
	}

	mediaId, err := s.WxApi.WebWxUploadMedia(s.WxWebCommon, s.WxWebXcg, s.Cookies, ss[len(ss)-1], b)
	if err != nil {
		glog.Error("[SendImage] Upload image failed, err = ", err)
		return -3, fmt.Errorf("[SendImage] Upload image failed, err = ", err)
	}

	ret, err := s.WxApi.WebWxSendMsgImg(s.WxWebCommon, s.WxWebXcg, s.Cookies, from, to, mediaId)
	if err != nil || ret != 0 {
		glog.Error("[SendImage] Send image failed, err = ", err)
		return ret, err
	} else {
		return ret, nil
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
			glog.Error("[LoginPolling] sec1 WebwxLogin failed, uuid = ", s.UuID, " err = [", err, "]")
			return -997
		} else if redirectUrl == "201;" {
			if flag == conf.SCAN {
				flag = conf.CONFIRM
			}

			glog.Info(">>> [LoginPolling] sec1 WebwxLogin uuid = ", s.UuID, " redirectUrl = ", redirectUrl)
			redirectUrl, err = s.WxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
			if err != nil {
				glog.Error("[LoginPolling] sec2 WebwxLogin failed, uuid = ", s.UuID, ", err = [", err, "]")
				s.UpdateLoginStat(-200)
				return -200
			} else if strings.Contains(redirectUrl, "http") {
				s.RedirectUrl = redirectUrl
				s.UpdateLoginStat(conf.LOGIN_SUCC)

				glog.Info(">>> [LoginPolling] sec2 WebwxLogin success, uuid = ", s.UuID)
				return 200
			}
		} else if redirectUrl == "408;" {
			s.UpdateLoginStat(conf.LOGIN_FAIL)
			if flag == conf.CONFIRM {
				flag = conf.SCAN
			}

			glog.Info(">>> [LoginPolling] sec3 WebwxLogin uuid = ", s.UuID, " redirectUrl = ", redirectUrl)
			tryCount++
			if tryCount >= conf.MAXTRY {
				s.UpdateLoginStat(-998)
				return -998
			} else {
				glog.Info(">>> [LoginPolling] sec3 WebwxLogin uuid = ", s.UuID, " retry =", tryCount)
			}
		} else if strings.Contains(redirectUrl, "http") {
			s.RedirectUrl = redirectUrl
			s.UpdateLoginStat(conf.LOGIN_SUCC)

			glog.Info(">>> [LoginPolling] sec3 WebwxLogin success, uuid = ", s.UuID)
			return 200
		} else {
			s.UpdateLoginStat(-999)

			glog.Error("[LoginPolling] sec4 WebwxLogin failed, uuid = ", s.UuID, ", redirectUrl = ", redirectUrl)
			return -999
		}
	}
}

func (s *Session) AnalizeVersion() {
	u, _ := url.Parse(s.RedirectUrl)

	// version may change
	s.WxWebCommon.Host = u.Host
	s.WxWebCommon.CgiDomain = u.Scheme + "://" + u.Host
	s.WxWebCommon.CgiUrl = s.WxWebCommon.CgiDomain + "/cgi-bin/mmwebwx-bin"

	s.WxWebCommon.SyncSrv = "webpush." + u.Host
	s.WxWebCommon.UploadUrl = "https://file." + u.Host + "/cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json"

	if strings.Contains(u.Host, "wx2") {
		// new version
		//s.WxWebCommon.SyncSrv = "webpush.wx2.qq.com"
		s.WxApi.Version = "wx2"
	} else {
		// old version
		//s.WxWebCommon.SyncSrv = "webpush.wx.qq.com"
		s.WxApi.Version = "wx"
	}
}

func (s *Session) InitUserCookies() int {
	var err error

	s.Cookies, err = s.WxApi.WebNewLoginPage(s.WxWebCommon, s.WxWebXcg, s.RedirectUrl)
	if err != nil {
		glog.Error("[InitUserCookies] WebNewLoginPage err = ", err)
		return -1
	}

	//获取用户临时会话列表
	session, err := s.WxApi.WebWxInit(s.WxWebCommon, s.WxWebXcg)
	if err != nil {
		glog.Error("[InitUserCookies] WebWxInit err = ", err)
		return -2
	}

	jc := &conf.JsonConfig{}
	jc, _ = conf.LoadJsonConfigFromBytes(session)

	s.SynKeyList, err = common.GetSyncKeyListFromJc(jc)
	if err != nil {
		glog.Error("[InitUserCookies] GetSyncKeyListFromJc err = ", err)
		return -3
	}

	s.Bot, err = common.GetUserInfoFromJc(jc)
	if err != nil {
		glog.Error("[InitUserCookies] GetUserInfoFromJc err = ", err)
		return -4
	}

	var contacts []byte
	contacts, err = s.WxApi.WebWxGetContact(s.WxWebCommon, s.WxWebXcg, s.Cookies)
	if err != nil {
		glog.Error("[InitUserCookies] WebWxGetContact err = ", err)
		return -5
	}

	s.ContactMgr, err = common.CreateContactManagerFromBytes(contacts)
	if err != nil {
		glog.Error("[InitUserCookies] CreateContactManagerFromBytes err = ", err)
		return -6
	}

	s.ContactMgr.AddContactFromUser(s.Bot)

	groups, err := common.GetSessionGroupFromJc(jc)
	if err != nil {
		glog.Error("[InitUserCookies] GetSessionGroupFromJc err = ", err)
		return -7
	}

	for _, group := range groups {
		s.ContactMgr.AddContactFromUser(group)
	}

	return 0
}

func (s *Session) Serve() {
	s.UpdateSrvStatus(true)
	glog.Info(">>> [Serve] Serve start userID = ", s.UserID)

	for s.Loop {
		//will be blocked here until wechat return response
		ret, selector, err := s.WxApi.SyncCheck(s.WxWebCommon, s.WxWebXcg, s.Cookies, s.WxWebCommon.SyncSrv, s.SynKeyList)
		if err != nil {
			glog.Error("[Serve] SyncCheck err = [", err, "] userID = ", s.UserID)
			continue
		}

		if ret == 0 {
			switch selector {
			case 2:
				//new message
				msg, err := s.WxApi.WebWxSync(s.WxWebCommon, s.WxWebXcg, s.Cookies, s.SynKeyList)
				if err != nil {
					glog.Error("[Serve] WebWxSync err = [", err, "] userID = ", s.UserID)
				} else {
					jc, err := conf.LoadJsonConfigFromBytes(msg)
					if err != nil {
						glog.Error("[Serve] LoadJsonConfigFromBytes err = [", err, "] userID = ", s.UserID)
						continue
					}

					msgCount, _ := jc.GetInt("AddMsgCount")
					if msgCount < 1 {
						continue
					}

					msgList, _ := jc.GetInterfaceSlice("AddMsgList")
					for _, v := range msgList {
						rmsg := s.Analize(v.(map[string]interface{}))
						glog.Info(">>> [Serve] FromUser = [", rmsg.FromUserName, "] Message Content = [", rmsg.Content, "] MessageType =[", rmsg.MsgType, "] UserID = [", s.UserID, "]")

						switch int(rmsg.MsgType) {
						case int(conf.MSG_TEXT):
							go s.SendReplyUserMessage(rmsg.FromUserName, rmsg.Content)
						case int(conf.MSG_SYS):
							go s.SendReplySysMessage(rmsg.FromUserName, rmsg.Content)
						}
					}
				}
			case 4:
				//session update
				_, err := s.WxApi.WebWxSync(s.WxWebCommon, s.WxWebXcg, s.Cookies, s.SynKeyList)
				if err != nil {
					glog.Error("[Serve] WebWxSync err = [", err, "] userID = ", s.UserID)
				} else {
					glog.Info(">>> [Serve] Session selector = ", selector, ", userID = ", s.UserID)
				}
			case 6:
				//contact update
				msg, err := s.WxApi.WebWxSync(s.WxWebCommon, s.WxWebXcg, s.Cookies, s.SynKeyList)
				if err != nil {
					glog.Error("[Serve] WebWxSync err = [", err, "] userID = ", s.UserID)
				} else {
					jc, err := conf.LoadJsonConfigFromBytes(msg)
					if err != nil {
						glog.Error("[Serve] LoadJsonConfigFromBytes err = [", err, "] userID = ", s.UserID)
						continue
					}

					count, _ := jc.GetInt("ModContactCount")
					if count < 1 {
						continue
					}

					newFriendList, _ := jc.GetInterfaceSlice("ModContactList")
					for _, newUser := range newFriendList {
						n := newUser.(map[string]interface{})
						u := &common.User{
							UserName: n["UserName"].(string),
							NickName: n["NickName"].(string),
						}

						exist := false
						for _, v := range s.ContactMgr.ContactList {
							if v.UserName == u.UserName {
								exist = true
								break
							}
						}

						if !exist {
							s.ContactMgr.AddContactFromUser(u)
						}

						go s.SendWelComeMessage(u.UserName)
					}
				}
			case 7:
				//session update
				_, err := s.WxApi.WebWxSync(s.WxWebCommon, s.WxWebXcg, s.Cookies, s.SynKeyList)
				if err != nil {
					glog.Error("[Serve] WebWxSync err = [", err, "] userID = ", s.UserID)
				} else {
					glog.Info(">>> [Serve] Session selector = ", selector, ", userID = ", s.UserID)
				}
			default:
				glog.Info(">>> [Serve] Session selector = ", selector, ", userID = ", s.UserID)
			}
		} else {
			// 1100: logout from client
			// 1101: login another webpage
			glog.Error("[Serve] Exception ret = ", ret, ", userID = ", s.UserID)
			break
		}
	}

	s.UpdateSrvStatus(false)
	glog.Info(">>> [Serve] Serve stop userID = ", s.UserID)
}

func (s *Session) Stop() {
	s.UpdateSrvStatus(false)
}

func (s *Session) UpdateSrvStatus(run bool) {
	s.serveLock.Lock()
	s.Loop = run
	s.serveLock.Unlock()
}

func (s *Session) GetSrvStatus() bool {
	s.serveLock.Lock()
	loop := s.Loop
	s.serveLock.Unlock()
	return loop
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

func (s *Session) SendWelComeMessage(userName string) {
Loop:
	for _, userConf := range s.AutoRepliesConf {
		if userConf.UserType == conf.USER_PERSON {
			for _, user := range s.ContactMgr.GetPersonContacts() {
				if user.UserName == userName {
					if userConf.WlmText != "" {
						var textMsg string
						if strings.Contains(userConf.WlmText, conf.WELCOME_USER_PATTEN) {
							textMsg = strings.Replace(userConf.WlmText, conf.WELCOME_USER_PATTEN, user.NickName, -1)
						} else {
							textMsg = userConf.WlmText
						}
						s.SendText(textMsg, s.Bot.UserName, userName)
						glog.Info(">>> [SendWelComeMessage] Auto reply from [", s.UserID, "] to [", user.NickName, "] text = [", textMsg, "]")
					}

					if userConf.WlmImage != "" {
						s.SendImage(userConf.WlmImage, s.Bot.UserName, userName)
						glog.Info(">>> [SendWelComeMessage] Auto reply from [", s.UserID, "] to [", user.NickName, "] image = [", userConf.WlmImage, "]")
					}
					break Loop
				}
			}
		}
	}
}

func (s *Session) SendReplyUserMessage(userName, msg string) {
	var userConf AutoReplyConf
	var toUser *common.User
	var contacts []*common.User
	match := false

	if strings.Contains(userName, "@@") {
		contacts = s.ContactMgr.GetGroupContacts()
	Loop1:
		for _, toUser = range contacts {
			if toUser.UserName == userName {
				glog.Info(">>> [ReplyUserMessage] UserName match = [", userName, "]")

				for _, userConf = range s.AutoRepliesConf {
					if userConf.NickName == toUser.NickName {
						glog.Info(">>> [ReplyUserMessage] NickName match = [", toUser.NickName, "]")
						match = true
						break Loop1
					}
				}
			}
		}
	} else {
		contacts = s.ContactMgr.GetPersonContacts()
	Loop2:
		for _, userConf = range s.AutoRepliesConf {
			if userConf.UserType == conf.USER_PERSON {
				for _, toUser = range contacts {
					if toUser.UserName == userName {
						glog.Info(">>> [ReplyUserMessage] UserName match = [", userName, "]")
						match = true
						break Loop2
					}
				}
			}
		}
	}

	if match {
		for _, keyword := range userConf.KeyWords {
			if strings.Contains(msg, keyword.Key) {
				glog.Info(">>> [ReplyUserMessage] KeyWord match = [", keyword.Key, "]")

				if keyword.Text != "" {
					s.SendText(keyword.Text, s.Bot.UserName, userName)
					glog.Info(">>> [ReplyUserMessage] Auto reply from [", s.UserID, "] to [", toUser.NickName, "] text = [", keyword.Text, "]")
				}

				if keyword.Image != "" {
					s.SendImage(keyword.Image, s.Bot.UserName, userName)
					glog.Info(">>> [ReplyUserMessage] Auto reply from [", s.UserID, "] to [", toUser.NickName, "] image = [", keyword.Image, "]")
				}
			}
		}
	}
}

func (s *Session) SendReplySysMessage(userName, message string) {
	for _, user := range s.ContactMgr.ContactList {
		if user.UserName == userName {
			newJoiner, err := utils.FindNewJoinerName(message)
			if err != nil {
				glog.Error("[ReplySysMessage] Get new joiner from system message failed, err = ", err)
			} else {
				go s.NewJoiner(user, newJoiner)
			}
			break
		}
	}
}

func (s *Session) NewJoiner(user *common.User, newJoiner string) {
	for _, vv := range s.AutoRepliesConf {
		if vv.NickName == user.NickName {
			if vv.WlmText != "" {
				welcome := strings.Replace(vv.WlmText, conf.WELCOME_USER_PATTEN, newJoiner, -1)
				s.SendText(welcome, s.Bot.UserName, user.UserName)
				glog.Info(">>> [NewJoiner] Auto reply from [", s.UserID, "] to [", user.NickName, "] text = [", welcome, "]")
			}

			if vv.WlmImage != "" {
				s.SendImage(vv.WlmImage, s.Bot.UserName, user.UserName)
				glog.Info(">>> [NewJoiner] Auto reply from [", s.UserID, "] to [", user.NickName, "] image = [", vv.WlmImage, "]")
			}
		}
	}
}

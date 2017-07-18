package module

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"service/common"
	"service/conf"
	"service/utils"
	"service/wxapi"
	"strings"
	"sync"
)

type KeyWord struct {
	key   string
	text  string
	image string
}

type AutoReplyConf struct {
	groupNickName    string
	wlmText  string
	wlmImage string
	keyWords []KeyWord
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
	autoRepliesConf []AutoReplyConf

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
		if glog.V(conf.LOG_LV) {
			glog.Error("[SendText] WebWxSendMsg Ret = ", ret, "ErrMsg = ", errMsg)
		}

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
		if glog.V(conf.LOG_LV) {
			glog.Error("[SendImage] Download image failed, err = ", err)
		}

		return -1, fmt.Errorf("[SendImage] Download image failed, err = ", err)
	}

	ss := strings.Split(fileName, "/")
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return -2, err
	}

	mediaId, err := s.WxApi.WebWxUploadMedia(s.WxWebCommon, s.WxWebXcg, s.cookies, ss[len(ss)-1], b)
	if err != nil {
		if glog.V(conf.LOG_LV) {
			glog.Error("[SendImage] Upload image failed")
		}

		return -3, fmt.Errorf("[SendImage] Upload image failed")
	}

	ret, err := s.WxApi.WebWxSendMsgImg(s.WxWebCommon, s.WxWebXcg, s.cookies, from, to, mediaId)
	if err != nil || ret != 0 {
		if glog.V(conf.LOG_LV) {
			glog.Error("[SendImage] Send image failed, err = ", err)
		}

		return ret, err
	} else {
		return ret, nil
	}
}

func (s *Session) InitSession(request *common.Msg_Create_Request) {
	if _, exist := SessionTable[s.UserID]; exist {
		if glog.V(conf.LOG_LV) {
			glog.Info(">>> [InitSession] Delete UserID ", s.UserID, "'s session")
		}
		delete(SessionTable, s.UserID)
	}

	SessionTable[s.UserID] = s
	s.autoRepliesConf = make([]AutoReplyConf, len(request.Config))

	for i := 0; i < len(request.Config); i++ {
		s.autoRepliesConf[i].groupNickName, _ = request.Config[i]["group"].(string)

		wlmText, exist := request.Config[i]["wlm_text"].(string)
		if exist {
			s.autoRepliesConf[i].wlmText = wlmText
		} else {
			s.autoRepliesConf[i].wlmText = ""
		}

		wlmImage, exist := request.Config[i]["wlm_image"].(string)
		if exist {
			s.autoRepliesConf[i].wlmImage = wlmImage
		} else {
			s.autoRepliesConf[i].wlmImage = ""
		}

		sections, exist := request.Config[i]["keywords"].([]interface{})
		if exist {
			s.autoRepliesConf[i].keyWords = make([]KeyWord, len(sections))

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

			if glog.V(conf.LOG_LV) {
				glog.Info(">>> [InitSession] Group [", s.autoRepliesConf[i].groupNickName, "] keyword configs = [", s.autoRepliesConf, "]")
			}
		} else {
			if glog.V(conf.LOG_LV) {
				glog.Info(">>> [InitSession] Group [", s.autoRepliesConf[i].groupNickName, "] has no keywords")
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
			if glog.V(conf.LOG_LV) {
				glog.Error("[LoginPolling] sec1 WebwxLogin failed, uuid = ", s.UuID, " err = [", err, "]")
			}
		} else if redirectUrl == "201;" {
			if flag == conf.SCAN {
				flag = conf.CONFIRM
			}

			if glog.V(conf.LOG_LV) {
				glog.Info(">>> [LoginPolling] sec1 WebwxLogin uuid = ", s.UuID, " redirectUrl = ", redirectUrl)
			}

			redirectUrl, err = s.WxApi.WebwxLogin(s.WxWebCommon, s.UuID, flag)
			if err != nil {
				if glog.V(conf.LOG_LV) {
					glog.Error("[LoginPolling] sec2 WebwxLogin failed, uuid = ", s.UuID, ", err = [", err, "]")
				}

				s.UpdateLoginStat(-200)
				return -200
			} else if strings.Contains(redirectUrl, "http") {
				s.RedirectUrl = redirectUrl
				s.UpdateLoginStat(conf.LOGIN_SUCC)

				if glog.V(conf.LOG_LV) {
					glog.Info(">>> [LoginPolling] sec2 WebwxLogin success, uuid = ", s.UuID)
				}

				return 200
			}
		} else if redirectUrl == "408;" {
			s.UpdateLoginStat(conf.LOGIN_FAIL)
			if flag == conf.CONFIRM {
				flag = conf.SCAN
			}

			if glog.V(conf.LOG_LV) {
				glog.Info(">>> [LoginPolling] sec3 WebwxLogin uuid = ", s.UuID, " redirectUrl = ", redirectUrl)
			}

			tryCount++
			if tryCount >= conf.MAXTRY {
				s.UpdateLoginStat(-998)
				return -998
			} else {
				if glog.V(conf.LOG_LV) {
					glog.Info(">>> [LoginPolling] sec3 WebwxLogin uuid = ", s.UuID, " retry =", tryCount)
				}
			}
		} else if strings.Contains(redirectUrl, "http") {
			s.RedirectUrl = redirectUrl
			s.UpdateLoginStat(conf.LOGIN_SUCC)

			if glog.V(conf.LOG_LV) {
				glog.Info(">>> [LoginPolling] sec3 WebwxLogin success, uuid = ", s.UuID)
			}

			return 200
		} else {
			s.UpdateLoginStat(-999)

			if glog.V(conf.LOG_LV) {
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
		if glog.V(conf.LOG_LV) {
			glog.Error("[InitUserCookies] WebNewLoginPage err = ", err)
		}
		return -1
	}

	//获取用户临时会话列表
	session, err := s.WxApi.WebWxInit(s.WxWebCommon, s.WxWebXcg)
	if err != nil {
		if glog.V(conf.LOG_LV) {
			glog.Error("[InitUserCookies] WebWxInit err = ", err)
		}
		return -2
	}

	jc := &conf.JsonConfig{}
	jc, _ = conf.LoadJsonConfigFromBytes(session)

	s.synKeyList, err = common.GetSyncKeyListFromJc(jc)
	if err != nil {
		if glog.V(conf.LOG_LV) {
			glog.Error("[InitUserCookies] GetSyncKeyListFromJc err = ", err)
		}
		return -3
	}

	s.Bot, err = common.GetUserInfoFromJc(jc)
	if err != nil {
		if glog.V(conf.LOG_LV) {
			glog.Error("[InitUserCookies] GetUserInfoFromJc err = ", err)
		}
		return -4
	}

	var contacts []byte
	contacts, err = s.WxApi.WebWxGetContact(s.WxWebCommon, s.WxWebXcg, s.cookies)
	if err != nil {
		if glog.V(conf.LOG_LV) {
			glog.Error("[InitUserCookies] WebWxGetContact err = ", err)
		}
		return -5
	}

	s.ContactMgr, err = common.CreateContactManagerFromBytes(contacts)
	if err != nil {
		if glog.V(conf.LOG_LV) {
			glog.Error("[InitUserCookies] CreateContactManagerFromBytes err = ", err)
		}
		return -6
	}

	s.ContactMgr.AddContactFromUser(s.Bot)

	groups, err := common.GetSessionGroupFromJc(jc)
	if err != nil {
		if glog.V(conf.LOG_LV) {
			glog.Error("[InitUserCookies] GetSessionGroupFromJc err = ", err)
		}
		return -7
	}

	for _, group := range groups {
		s.ContactMgr.AddContactFromUser(group)
	}

	return 0
}

func (s *Session) Serve() {
	if glog.V(conf.LOG_LV) {
		glog.Info(">>> [Serve] Looping start userID = ", s.UserID)
	}

	for s.Loop {
		//Will be blocked here until wechat return response
		ret, selector, err := s.WxApi.SyncCheck(s.WxWebCommon, s.WxWebXcg, s.cookies, s.WxWebCommon.SyncSrv, s.synKeyList)
		if err != nil {
			if glog.V(conf.LOG_LV) {
				glog.Error("[Serve] SyncCheck err = [", err, "] userID = ", s.UserID)
			}
			continue
		}

		if ret == 0 {
			//if glog.V(conf.LOG_LV) {
			//	glog.Info(">>> [Serve] SyncCheck new message get, ret = ", ret, " && selector = ", selector, " userID = ", s.UserID)
			//}

			if selector == 2 {
				msg, err := s.WxApi.WebWxSync(s.WxWebCommon, s.WxWebXcg, s.cookies, s.synKeyList)
				if err != nil {
					if glog.V(conf.LOG_LV) {
						glog.Error("[Serve] WebWxSync err = [", err, "] userID = ", s.UserID)
					}
				} else {
					jc, err := conf.LoadJsonConfigFromBytes(msg)
					if err != nil {
						if glog.V(conf.LOG_LV) {
							glog.Error("[Serve] LoadJsonConfigFromBytes err = [", err, "] userID = ", s.UserID)
						}
						continue
					}

					msgCount, _ := jc.GetInt("AddMsgCount")
					if msgCount < 1 {
						continue
					}

					msgis, _ := jc.GetInterfaceSlice("AddMsgList")
					for _, v := range msgis {
						rmsg := s.Analize(v.(map[string]interface{}))
						if glog.V(conf.LOG_LV) {
							glog.Info(">>> [Serve] FromUser = [", rmsg.FromUserName, "] Message Content = [", rmsg.Content, "] MessageType =[", rmsg.MsgType, "] UserID = [", s.UserID, "]")
						}

						switch int(rmsg.MsgType) {
						case int(conf.MSG_TEXT):
							go s.ReplyUserMessage(rmsg.FromUserName, rmsg.Content)
						case int(conf.MSG_SYS):
							go s.ReplySysMessage(rmsg.FromUserName, rmsg.Content)
						}
					}
				}
			} else if selector != 0 && selector != 7 {
				if glog.V(conf.LOG_LV) {
					glog.Error("[Serve] Session down, selector = ", selector, ", userID = ", s.UserID)
				}
				break
			}
		} else if ret == 1101 || ret == 1100 {
			if glog.V(conf.LOG_LV) {
				glog.Error("[Serve] User logout error code = ", ret, ", userID = ", s.UserID)
			}
			break
		} else if ret == 1205 {
			if glog.V(conf.LOG_LV) {
				glog.Error("[Serve] Api blocked, ret = ", 1205, ", userID = ", s.UserID)
			}
			break
		} else {
			if glog.V(conf.LOG_LV) {
				glog.Error("[Serve] Unhandled exception ret = ", ret, ", userID = ", s.UserID)
			}
			break
		}
	}

	if glog.V(conf.LOG_LV) {
		glog.Info(">>> [Serve] Looping stop userID = ", s.UserID)
	}
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

func (s *Session) ReplyUserMessage(group, msg string) {
	for _, toUser := range s.ContactMgr.ContactList {
		if toUser.UserName == group {
			for _, groupConf := range s.autoRepliesConf {
				if groupConf.groupNickName == toUser.NickName {
					for _, keyword := range groupConf.keyWords {
						if strings.Contains(msg, keyword.key) {
							if keyword.text != "" {
								go s.SendText(keyword.text, s.Bot.UserName, toUser.UserName)
							}

							if keyword.image != "" {
								go s.SendImage(keyword.image, s.Bot.UserName, toUser.UserName)
							}
						}
					}
				}
			}
			return
		}
	}
}

func (s *Session) ReplySysMessage(userName, msg string) {
	reg := regexp.MustCompile(conf.NEW_JOINER_PATTERN)
	mstr := reg.FindString(msg)
	//welcome := fmt.Sprintf(conf.WELCOME_MESSAGE, strings.Replace(mstr, "\"", "", -1))

	for _, v := range s.ContactMgr.ContactList {
		if v.UserName == userName {
			for _, vv := range s.autoRepliesConf {
				if vv.groupNickName == v.NickName {
					if vv.wlmText != "" {
						welcome := strings.Replace(vv.wlmText, conf.WELCOME_USER_PATTEN, mstr, -1)
						//welcome := strings.Replace(vv.wlmText, conf.WELCOME_USER_PATTEN, strings.Replace(mstr, "\"", "", -1), -1)
						go s.SendText(welcome, s.Bot.UserName, userName)
					}

					if vv.wlmImage != "" {
						go s.SendImage(vv.wlmImage, s.Bot.UserName, userName)
					}
				}
			}
		}
	}
}

func (s *Session) Stop() {
	s.Loop = false
}

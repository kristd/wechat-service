package common

import (
	"service/utils"
)

type Common struct {
	AppId     string
	Fun       string
	LoginUrl  string
	Lang      string
	DeviceID  string
	UserAgent string
	//Diff version
	Host        string
	CgiUrl      string
	CgiDomain   string
	SyncSrv     string
	UploadUrl   string
	MediaCount  uint32
	RedirectUri string
}

var (
	// DefaultCommon: default session config
	DefaultCommon = &Common{
		AppId:      "wx782c26e4c19acffb",
		Fun:        "new",
		LoginUrl:   "https://login.weixin.qq.com",
		Lang:       "zh_CN",
		DeviceID:   "e" + utils.GetRandomStringFromNum(15),
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36",
		CgiUrl:     "https://wx.qq.com/cgi-bin/mmwebwx-bin",
		SyncSrv:    "webpush.wx.qq.com",
		UploadUrl:  "https://file.wx.qq.com/cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json",
		MediaCount: 0,
	}
)

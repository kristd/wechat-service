package conf

const (
	API_CREATE = "/api/create"
	API_LOGIN  = "/api/login"
	API_SEND   = "/api/send"
	API_EXIT   = "/api/exit"

	MAX_BUF_SIZE = 204800

	SCAN    = "1"
	CONFIRM = "0"

	LOGIN_WAIT = 201
	LOGIN_SUCC = 200
	LOGIN_FAIL = 408

	// msg types
	MSG_TEXT        = 1     // text message
	MSG_IMG         = 3     // image message
	MSG_VOICE       = 34    // voice message
	MSG_FV          = 37    // friend verification message
	MSG_PF          = 40    // POSSIBLEFRIEND_MSG
	MSG_SCC         = 42    // shared contact card
	MSG_VIDEO       = 43    // video message
	MSG_EMOTION     = 47    // gif
	MSG_LOCATION    = 48    // location message
	MSG_LINK        = 49    // shared link message
	MSG_VOIP        = 50    // VOIPMSG
	MSG_INIT        = 51    // wechat init message
	MSG_VOIPNOTIFY  = 52    // VOIPNOTIFY
	MSG_VOIPINVITE  = 53    // VOIPINVITE
	MSG_SHORT_VIDEO = 62    // short video message
	MSG_SYSNOTICE   = 9999  // SYSNOTICE
	MSG_SYS         = 10000 // system message
	MSG_WITHDRAW    = 10002 // withdraw notification message

	CLIENT_CREATE = 1
	CLIENT_LOGIN  = 2
	CLIENT_BACK   = 3
	CLIENT_SEND   = 4
	CLIENT_EXIT   = 5
	CLIENT_BEAT   = 6

	TEXT_MSG = 1
	IMG_MSG  = 2

	MAXTRY = 5

	NEW_JOINER_PATTERN  = `"(.*?)"`
	WELCOME_USER_PATTEN = "${username}"

	LOG_LV = 2

	HEARTBEAT_INTERVAL = 60

	USER_GROUP  = 101
	USER_PERSON = 102

	INTERVAL          = 60
	WECHAT_FILEHELPER = "filehelper"
)

var (
	/*
		扫码加入群/第三方拉入群/群主拉入群/游戏中心进入群
			"李晓云"邀请"TJ"加入了群聊 || "TJ" invited "李晓云" to the group chat
			"TJ"通过扫描你分享的二维码加入了群聊 || "李晓云" joined the group chat via your shared QR Code
			"李晓云"通过扫描"TJ"分享的二维码加入了群聊 || "李晓云" joined the group chat via the QR Code shared by "TJ"
			你邀请"李晓云"加入了群聊 || You've invited "Oscar" to the group chat
			李晓云通过游戏中心加入群聊
	*/
	WELCOME_MESSAGE_PATTERN = [...]string{`邀请(.+)加入了群聊`, `(.+)通过扫描`, `invited (.+) to the`, `(.+) joined the group`, `(.+)通过游戏中心加入群聊`}
)

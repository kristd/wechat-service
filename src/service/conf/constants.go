package conf

const (
	MAX_BUF_SIZE = 2048

	IMG_SAVE_PATH = "./image/"

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

	TEXT_MSG = 1
	IMG_MSG  = 2

	MAXTRY = 5

	NEW_JOINER_PATTERN = `\"(.*?)\"`
	WELCOME_USER_PATTEN = "${username}"

	LOG_LV = 2
)

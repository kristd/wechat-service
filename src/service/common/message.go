package common


type Msg_Create_Request struct {
	Action int                      `json:"action"`
	Config []map[string]interface{} `json:"conf"`
	UserID int                      `json:"id"`
}

type Msg_Create_Response struct {
	Action int    `json:"action"`
	UserID int    `json:"id"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Uuid   string `json:"uuid"`
	QrCode string `json:"qrcode"`
}

type Msg_Login_Request struct {
	Action int `json:"action"`
	UserID int `json:"id"`
}

type Msg_Login_Response struct {
	Action int    `json:"action"`
	UserID int    `json:"id"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}

type Msg_Send_Request struct {
	Action int    `json:"action"`
	UserID int    `json:"id"`
	Group  string `json:"group"`
	Params struct {
		Type    int    `json:"type"`
		Method  string `json:"method"`
		Content string `json:"content"`
	} `json:"params"`
}

type Msg_Send_Response struct {
	Action int    `json:"action"`
	UserID int    `json:"id"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
}

type Msg_Exit_Request struct {
	Action int `json:"action"`
	UserID int `json:"id"`
}

type Msg_Exit_Response struct {
	Action int    `json:"action"`
	UserID int    `json:"id"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
}

// BaseRequest: http request body BaseRequest
type BaseRequest struct {
	Uin      string
	Sid      string
	Skey     string
	DeviceID string
}

// BaseResponse: web api http response body BaseResponse struct
type BaseResponse struct {
	Ret    int
	ErrMsg string
}

// InitReqBody: common http request body struct
type InitReqBody struct {
    BaseRequest        *BaseRequest
    Msg                interface{}
    SyncKey            *SyncKeyList
    Rr                 int
    Code               int
    FromUserName       string
    ToUserName         string
    ClientMsgId        int
    ClientMediaId      int
    TotalLen           int
    StartPos           int
    DataLen            int
    MediaType          int
    Scene              int
    Count              int
    List               []*User
    Opcode             int
    SceneList          []int
    SceneListCount     int
    VerifyContent      string
    VerifyUserList     []*VerifyUser
    VerifyUserListSize int
    skey               string
    MemberCount        int
    MemberList         []*User
    Topic              string
}

// TextMessage: text message struct
type TextMessage struct {
    Type         int
    Content      string
    FromUserName string
    ToUserName   string
    LocalID      int
    ClientMsgId  int
}

// MediaMessage
type MediaMessage struct {
    Type         int
    Content      string
    FromUserName string
    ToUserName   string
    LocalID      int
    ClientMsgId  int
    MediaId      string
}

type RecommendInfo struct {
    Ticket   string
    UserName string
    NickName string
    Content  string
    Sex      int
}

// ReceivedMessage: for received message
type ReceivedMessage struct {
    IsGroup       bool
    MsgId         string
    Content       string
    FromUserName  string
    ToUserName    string
    Who           string
    MsgType       int
    OriginContent string
    At            string

    RecommendInfo *RecommendInfo
}
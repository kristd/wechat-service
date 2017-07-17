package common

import (
    "service/conf"
    "reflect"
)

// User: contact struct
type User struct {
	Uin               int
	UserName          string
	NickName          string
	HeadImgUrl        string
	ContactFlag       int
	MemberCount       int
	MemberList        []*User
	RemarkName        string
	PYInitial         string
	PYQuanPin         string
	RemarkPYInitial   string
	RemarkPYQuanPin   string
	HideInputBarFlag  int
	StarFriend        int
	Sex               int
	Signature         string
	AppAccountFlag    int
	Statues           int
	AttrStatus        uint32
	Province          string
	City              string
	Alias             string
	VerifyFlag        int
	OwnerUin          int
	WebWxPluginSwitch int
	HeadImgFlag       int
	SnsFlag           int
	UniFriend         int
	DisplayName       string
	ChatRoomId        int
	KeyWord           string
	EncryChatRoomId   string
	IsOwner           int
	MemberStatus      int
}

// VerifyUser: verify user request body struct
type VerifyUser struct {
	Value            string
	VerifyUserTicket string
}

func GetUserInfoFromJc(jc *conf.JsonConfig) (*User, error) {
    user, _ := jc.GetInterface("User")
    u := &User{}
    fields := reflect.ValueOf(u).Elem()
    for k, v := range user.(map[string]interface{}) {
        field := fields.FieldByName(k)
        if vv, ok := v.(float64); ok {
            field.Set(reflect.ValueOf(int(vv)))
        } else {
            field.Set(reflect.ValueOf(v))
        }
    }
    return u, nil
}

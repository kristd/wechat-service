package common

import (
	"encoding/json"
	"strings"
)

// ContactManager: contact manager
type ContactManager struct {
	ContactList []*User //contact list
}

// ContactResponse: get contact response struct
type ContactResponse struct {
	BaseResponse *BaseResponse
	MemberCount  int
	MemberList   []*User
	Seq          int
}

var (
	// SpecialContact: special contacts map
	SpecialContact = map[string]bool{
		"filehelper":            true,
		"newsapp":               true,
		"fmessage":              true,
		"weibo":                 true,
		"qqmail":                true,
		"tmessage":              true,
		"qmessage":              true,
		"qqsync":                true,
		"floatbottle":           true,
		"lbsapp":                true,
		"shakeapp":              true,
		"medianote":             true,
		"qqfriend":              true,
		"readerapp":             true,
		"blogapp":               true,
		"facebookapp":           true,
		"masssendapp":           true,
		"meishiapp":             true,
		"feedsapp":              true,
		"voip":                  true,
		"blogappweixin":         true,
		"weixin":                true,
		"brandsessionholder":    true,
		"weixinreminder":        true,
		"officialaccounts":      true,
		"wxitil":                true,
		"userexperience_alarm":  true,
		"notification_messages": true,
	}
)

// CreateContactManagerFromBytes: create contact maanger from bytes
func CreateContactManagerFromBytes(cb []byte) (*ContactManager, error) {
	var cr ContactResponse
	if err := json.Unmarshal(cb, &cr); err != nil {
		return nil, err
	}
	cm := &ContactManager{
		ContactList: cr.MemberList,
	}
	return cm, nil
}

// AddContactFromBytes
// upate contact manager from bytes
func (s *ContactManager) AddContactFromBytes(cb []byte) error {
	var cr ContactResponse
	if err := json.Unmarshal(cb, &cr); err != nil {
		return err
	}
	s.ContactList = append(s.ContactList, cr.MemberList...)
	return nil
}

// AddContactFromUser
// add a new user contact to contact manager
func (s *ContactManager) AddContactFromUser(user *User) {
	if user == nil {
		return
	}
	s.ContactList = append(s.ContactList, user)
}

// GetContactByUserName
// get contact by UserName
func (s *ContactManager) GetContactByUserName(un string) *User {
	for _, v := range s.ContactList {
		if v.UserName == un {
			return v
		}
	}
	return nil
}

// GetGroupContacts: get group contacts
func (s *ContactManager) GetGroupContacts() []*User {
	clarray := make([]*User, 0)
	for _, v := range s.ContactList {
		if strings.Contains(v.UserName, "@@") {
			clarray = append(clarray, v)
		}
	}
	return clarray
}

// GetStrangers: not group and not StarFriend
func (s *ContactManager) GetStrangers() []*User {
	clarray := make([]*User, 0)
	for _, v := range s.ContactList {
		if !strings.Contains(v.UserName, "@@") &&
			v.StarFriend == 0 &&
			v.VerifyFlag&8 == 0 &&
			!SpecialContact[v.UserName] {
			clarray = append(clarray, v)
		}
	}
	return clarray
}

// GetContactByName: get contacts by name
func (s *ContactManager) GetContactByName(sig string) []*User {
	clarray := make([]*User, 0)
	for _, v := range s.ContactList {
		if v.NickName == sig || v.RemarkName == sig {
			clarray = append(clarray, v)
		}
	}
	return clarray
}

// GetContactByPYQuanPin: get contact by User.PYQuanPin
func (s *ContactManager) GetContactByPYQuanPin(sig string) *User {
	for _, v := range s.ContactList {
		if v.PYQuanPin == sig || v.RemarkPYQuanPin == sig {
			return v
		}
	}
	return nil
}

// GetAll: get all contacts
func (s *ContactManager) GetAll() []*User {
	return s.ContactList
}

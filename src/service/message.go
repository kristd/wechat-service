package main

import (

)

const (
	Client_Action_Create = 1
	Client_Action_Login  = 2
    Client_Action_Back   = 3
	Client_Action_Send   = 4
	Client_Action_Exit   = 5
)

const (
    TEXT_MSG    = 1
    IMG_MSG     = 2
)


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
    Action int      `json:"action"`
    UserID int      `json:"id"`
}

type Msg_Login_Response struct {
    Action  int         `json:"action"`
    UserID  int         `json:"id"`
    Code    int         `json:"code"`
    Msg     string      `json:"msg"`
    Status  int         `json:"status"`
}

type Msg_Send_Request struct {
    Action int      `json:"action"`
    UserID int      `json:"id"`
    Group  string   `json:"group"`
    Params struct {
        Type        int         `json:"type"`
        Method      string      `json:"method"`
        Content     string      `json:"content"`
    } `json:"params"`
}

type Msg_Send_Response struct {
    Action  int         `json:"action"`
    UserID  int         `json:"id"`
    Code    int         `json:"code"`
    Msg     string      `json:"msg"`
}

type Msg_Exit_Request struct {
    Action int      `json:"action"`
    UserID int      `json:"id"`
}

type Msg_Exit_Response struct {
    Action  int         `json:"action"`
    UserID  int         `json:"id"`
    Code    int         `json:"code"`
    Msg     string      `json:"msg"`
}









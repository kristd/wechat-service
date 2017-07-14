/*
    https://github.com/songtianyi/wechat-go
*/

package main

import (
    "net/url"
    "net/http"
    "fmt"
    "io/ioutil"
    "strings"
    "strconv"
    "time"
    "encoding/xml"
    "encoding/json"
    "bytes"
    "net/http/cookiejar"
    "mime/multipart"
    "io"
    "sync/atomic"
    "regexp"
)


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

// SyncKey: struct for synccheck
type SyncKey struct {
    Key int
    Val int
}

// SyncKeyList: list of synckey
type SyncKeyList struct {
    Count int
    List  []SyncKey
}


// InitReqBody: common http request body struct
type InitReqBody struct {
    BaseRequest        *BaseRequest
    Msg                interface{}
    SyncKey            *SyncKeyList
    rr                 int
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

type WebwxApi struct {

}

// s.String output synckey list in string
func (s *SyncKeyList) String() string {
    strs := make([]string, 0)

    //fmt.Println(">>> s.List <<< =", s.List)

    for _, v := range s.List {
        strs = append(strs, strconv.Itoa(v.Key)+"_"+strconv.Itoa(v.Val))
    }
    return strings.Join(strs, "|")
}

func (wx *WebwxApi) WebwxGetUuid(common *Common) (uuid string, qrcode string) {
    params := url.Values{}
    params.Add("appid", common.AppId)
    params.Add("fun", common.Fun)
    params.Add("lang", common.Lang)
    params.Add("_", GetTimeStamp(10))
    addrUrl := common.LoginUrl + "/jslogin?" + params.Encode()

    req, err := http.NewRequest("GET", addrUrl, nil)
    if err != nil {
        fmt.Println("http.NewRequest err =", err)
    }

    req.Header.Add("User-Agent", common.UserAgent)

    client := http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("client.Do err =", err)
    }

    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)

    uuid = strings.Split(string(body), "\"")[1]
    qrcode = common.LoginUrl + "/qrcode/" + uuid

    return uuid, qrcode
}


func (wx *WebwxApi) WebwxLogin(common *Common, uuid string, tip string) (string, error) {
    params := url.Values{}
    params.Add("tip", tip)
    params.Add("uuid", uuid)
    params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
    params.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
    uri := common.LoginUrl + "/cgi-bin/mmwebwx-bin/login?" + params.Encode()
    resp, err := http.Get(uri)
    if err != nil {
        return "", err
    }

    r := &http.Request{}
    r.Context()

    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    strb := string(body)
    if strings.Contains(strb, "window.code=200") && strings.Contains(strb, "window.redirect_uri") {
        ss := strings.Split(strb, "\"")
        if len(ss) < 2 {
            return "", fmt.Errorf("parse redirect_uri fail, %s", strb)
        }
        return ss[1], nil
    } else if strings.Contains(strb, "window.code=201") {
        ss := strings.Split(strb, "=")
        return ss[1], nil
    } else if strings.Contains(strb, "window.code=408") {
        ss := strings.Split(strb, "=")
        return ss[1], nil
    } else {
        return "", fmt.Errorf("invalid response, %s", strb)
    }
}


// WebNewLoginPage: webwxnewloginpage api
func (wx *WebwxApi) WebNewLoginPage(common *Common, xc *XmlConfig, uri string) ([]*http.Cookie, error) {
    u, _ := url.Parse(uri)
    km := u.Query()
    km.Add("fun", "new")
    uri = common.CgiUrl + "/webwxnewloginpage?" + km.Encode()
    //uri = "https://wx.qq.com/cgi-bin/mmwebwx-bin" + "/webwxnewloginpage?" + km.Encode()
    resp, err := http.Get(uri)
    if err != nil {
        fmt.Println("http.Get err =", err, " url =", uri)
        return nil, err
    }

    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)

    fmt.Println("ioutil.ReadAll =", string(body))

    if err := xml.Unmarshal(body, xc); err != nil {
        return nil, err
    }
    if xc.Ret != 0 {
        return nil, fmt.Errorf("xc.Ret != 0, %s", string(body))
    }

    fmt.Println("xml config =", xc)
    fmt.Println("response body =", resp.Body)

    return resp.Cookies(), nil
}


// WebWxInit: 获取会话列表
func (wx *WebwxApi) WebWxInit(common *Common, ce *XmlConfig) ([]byte, error) {
    km := url.Values{}
    km.Add("pass_ticket", ce.PassTicket)
    km.Add("skey", ce.Skey)
    km.Add("r", strconv.FormatInt(time.Now().Unix(), 10))

    uri := common.CgiUrl + "/webwxinit?" + km.Encode()

    js := InitReqBody {
        BaseRequest: &BaseRequest{
            ce.Wxuin,
            ce.Wxsid,
            ce.Skey,
            common.DeviceID,
        },
    }

    b, _ := json.Marshal(js)
    client := &http.Client{}
    req, err := http.NewRequest("POST", uri, bytes.NewReader(b))
    req.Header.Add("Content-Type", "application/json; charset=UTF-8")
    req.Header.Add("User-Agent", common.UserAgent)

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    return body, nil
}


// WebWxGetContact: 获取联系人列表
func (wx *WebwxApi) WebWxGetContact(common *Common, ce *XmlConfig, cookies []*http.Cookie) ([]byte, error) {
    km := url.Values{}
    km.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
    km.Add("seq", "0")
    km.Add("skey", ce.Skey)
    uri := common.CgiUrl + "/webwxgetcontact?" + km.Encode()

    js := InitReqBody{
        BaseRequest: &BaseRequest{
            ce.Wxuin,
            ce.Wxsid,
            ce.Skey,
            common.DeviceID,
        },
    }

    b, _ := json.Marshal(js)
    req, err := http.NewRequest("POST", uri, bytes.NewReader(b))
    if err != nil {
        return nil, err
    }
    req.Header.Add("Content-Type", "application/json; charset=UTF-8")
    req.Header.Add("User-Agent", common.UserAgent)

    jar, _ := cookiejar.New(nil)
    u, _ := url.Parse(uri)
    jar.SetCookies(u, cookies)
    client := &http.Client{Jar: jar}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    return body, nil
}

// WebWxSendMsg: webwxsendmsg api
func (wx *WebwxApi) WebWxSendMsg(common *Common, ce *XmlConfig, cookies []*http.Cookie, from, to string, msg string) ([]byte, error) {
    km := url.Values{}
    km.Add("pass_ticket", ce.PassTicket)
    uri := common.CgiUrl + "/webwxsendmsg?" + km.Encode()

    js := InitReqBody{
        BaseRequest: &BaseRequest{
            ce.Wxuin,
            ce.Wxsid,
            ce.Skey,
            common.DeviceID,
        },
        Msg: &TextMessage{
            Type:         1,
            Content:      msg,
            FromUserName: from,
            ToUserName:   to,
            LocalID:      int(time.Now().Unix() * 1e4),
            ClientMsgId:  int(time.Now().Unix() * 1e4),
        },
    }

    b, _ := json.Marshal(js)

    fmt.Println("")
    fmt.Println(">>>>> WebWxSendMsg req <<<<<", string(b))

    req, err := http.NewRequest("POST", uri, bytes.NewReader(b))
    if err != nil {
        return nil, err
    }
    req.Header.Add("Content-Type", "application/json; charset=UTF-8")
    req.Header.Add("User-Agent", common.UserAgent)

    jar, _ := cookiejar.New(nil)
    u, _ := url.Parse(uri)
    jar.SetCookies(u, cookies)
    client := &http.Client{Jar: jar}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    return body, nil
}


// WebWxUploadMedia: webwxuploadmedia api
func (wx *WebwxApi) WebWxUploadMedia(common *Common, ce *XmlConfig, cookies []*http.Cookie, filename string, content []byte) (string, error) {
    var b bytes.Buffer
    w := multipart.NewWriter(&b)
    fw, _ := w.CreateFormFile("filename", filename)
    if _, err := io.Copy(fw, bytes.NewReader(content)); err != nil {
        return "", err
    }

    ss := strings.Split(filename, ".")
    if len(ss) != 2 {
        return "", fmt.Errorf("file type suffix not found")
    }
    suffix := ss[1]

    fw, _ = w.CreateFormField("id")
    _, _ = fw.Write([]byte("WU_FILE_" + strconv.Itoa(int(common.MediaCount))))
    common.MediaCount = atomic.AddUint32(&common.MediaCount, 1)

    fw, _ = w.CreateFormField("name")
    _, _ = fw.Write([]byte(filename))

    fw, _ = w.CreateFormField("type")
    if suffix == "gif" {
        _, _ = fw.Write([]byte("image/gif"))
    } else {
        _, _ = fw.Write([]byte("image/jpeg"))
    }

    fw, _ = w.CreateFormField("lastModifieDate")
    _, _ = fw.Write([]byte("Mon Feb 13 2017 17:27:23 GMT+0800 (CST)"))

    fw, _ = w.CreateFormField("size")
    _, _ = fw.Write([]byte(strconv.Itoa(len(content))))

    fw, _ = w.CreateFormField("mediatype")
    if suffix == "gif" {
        _, _ = fw.Write([]byte("doc"))
    } else {
        _, _ = fw.Write([]byte("pic"))
    }

    js := InitReqBody{
        BaseRequest: &BaseRequest{
            ce.Wxuin,
            ce.Wxsid,
            ce.Skey,
            common.DeviceID,
        },
        ClientMediaId: int(time.Now().Unix() * 1e4),
        TotalLen:      len(content),
        StartPos:      0,
        DataLen:       len(content),
        MediaType:     4,
    }

    jb, _ := json.Marshal(js)

    fw, _ = w.CreateFormField("uploadmediarequest")
    _, _ = fw.Write(jb)

    fw, _ = w.CreateFormField("webwx_data_ticket")
    for _, v := range cookies {
        if strings.Contains(v.String(), "webwx_data_ticket") {
            _, _ = fw.Write([]byte(strings.Split(v.String(), "=")[1]))
            break
        }
    }

    fw, _ = w.CreateFormField("pass_ticket")
    _, _ = fw.Write([]byte(ce.PassTicket))
    w.Close()

    req, err := http.NewRequest("POST", common.UploadUrl, &b)
    if err != nil {
        return "", err
    }
    req.Header.Add("Content-Type", w.FormDataContentType())
    req.Header.Add("User-Agent", common.UserAgent)

    jar, _ := cookiejar.New(nil)
    u, _ := url.Parse(common.UploadUrl)
    jar.SetCookies(u, cookies)
    client := &http.Client{Jar: jar}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    jc, err := LoadJsonConfigFromBytes(body)
    if err != nil {
        return "", err
    }
    ret, _ := jc.GetInt("BaseResponse.Ret")
    if ret != 0 {
        return "", fmt.Errorf("BaseResponse.Ret=%d", ret)
    }
    mediaId, _ := jc.GetString("MediaId")
    return mediaId, nil
}


// WebWxSendMsgImg: webwxsendmsgimg api
func (wx *WebwxApi) WebWxSendMsgImg(common *Common, ce *XmlConfig, cookies []*http.Cookie,
    from, to, media string) (int, error) {

    km := url.Values{}
    km.Add("pass_ticket", ce.PassTicket)
    km.Add("fun", "async")
    km.Add("f", "json")
    km.Add("lang", common.Lang)

    uri := common.CgiUrl + "/webwxsendmsgimg?" + km.Encode()

    js := InitReqBody{
        BaseRequest: &BaseRequest{
            ce.Wxuin,
            ce.Wxsid,
            ce.Skey,
            common.DeviceID,
        },
        Msg: &MediaMessage{
            Type:         3,
            Content:      "",
            FromUserName: from,
            ToUserName:   to,
            LocalID:      int(time.Now().Unix() * 1e4),
            ClientMsgId:  int(time.Now().Unix() * 1e4),
            MediaId:      media,
        },
        Scene: 0,
    }

    b, _ := json.Marshal(js)
    req, err := http.NewRequest("POST", uri, bytes.NewReader(b))
    if err != nil {
        return -1, err
    }
    req.Header.Add("Content-Type", "application/json; charset=UTF-8")
    req.Header.Add("User-Agent", common.UserAgent)

    jar, _ := cookiejar.New(nil)
    u, _ := url.Parse(uri)
    jar.SetCookies(u, cookies)
    client := &http.Client{Jar: jar}
    resp, err := client.Do(req)
    if err != nil {
        return -1, err
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    jc, _ := LoadJsonConfigFromBytes(body)
    ret, _ := jc.GetInt("BaseResponse.Ret")
    return ret, nil
}


// SyncCheck: synccheck api - get all the new message
func (wx *WebwxApi) SyncCheck(common *Common, ce *XmlConfig, cookies []*http.Cookie,  server string, skl *SyncKeyList) (int, int, error) {
    km := url.Values{}
    km.Add("r", strconv.FormatInt(time.Now().Unix()*1000, 10))
    km.Add("sid", ce.Wxsid)
    km.Add("uin", ce.Wxuin)
    km.Add("skey", ce.Skey)
    km.Add("deviceid", common.DeviceID)
    km.Add("synckey", skl.String())
    km.Add("_", strconv.FormatInt(time.Now().Unix()*1000, 10))
    uri := "https://" + server + "/cgi-bin/mmwebwx-bin/synccheck?" + km.Encode()

    js := InitReqBody{
        BaseRequest: &BaseRequest{
            ce.Wxuin,
            ce.Wxsid,
            ce.Skey,
            common.DeviceID,
        },
    }

    b, _ := json.Marshal(js)
    jar, _ := cookiejar.New(nil)
    u, _ := url.Parse(uri)
    jar.SetCookies(u, cookies)
    client := &http.Client{Jar: jar, Timeout: time.Duration(30) * time.Second}
    req, err := http.NewRequest("GET", uri, bytes.NewReader(b))
    if err != nil {
        return 0, 0, err
    }

    req.Header.Add("Content-Type", "application/json; charset=UTF-8")
    req.Header.Add("User-Agent", common.UserAgent)

    resp, err := client.Do(req)
    if err != nil {
        return 0, 0, err
    }

    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    strb := string(body)
    reg := regexp.MustCompile("window.synccheck={retcode:\"(\\d+)\",selector:\"(\\d+)\"}")
    sub := reg.FindStringSubmatch(strb)
    retcode, _ := strconv.Atoi(sub[1])
    selector, _ := strconv.Atoi(sub[2])
    return retcode, selector, nil
}


// WebWxSync: webwxsync api
func (wx *WebwxApi) WebWxSync(common *Common, ce *XmlConfig, cookies []*http.Cookie, skl *SyncKeyList) ([]byte, error) {

    km := url.Values{}
    km.Add("skey", ce.Skey)
    km.Add("sid", ce.Wxsid)
    km.Add("lang", common.Lang)
    km.Add("pass_ticket", ce.PassTicket)

    uri := common.CgiUrl + "/webwxsync?" + km.Encode()

    js := InitReqBody{
        BaseRequest: &BaseRequest{
            ce.Wxuin,
            ce.Wxsid,
            ce.Skey,
            common.DeviceID,
        },
        SyncKey: skl,
        rr:      ^int(time.Now().Unix()) + 1,
    }

    b, _ := json.Marshal(js)
    jar, _ := cookiejar.New(nil)
    u, _ := url.Parse(uri)
    jar.SetCookies(u, cookies)
    client := &http.Client{Jar: jar, Timeout: time.Duration(10) * time.Second}
    req, err := http.NewRequest("POST", uri, bytes.NewReader(b))
    req.Header.Add("Content-Type", "application/json; charset=UTF-8")
    req.Header.Add("User-Agent", common.UserAgent)

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)

    jc, err := LoadJsonConfigFromBytes(body)
    if err != nil {
        return nil, err
    }
    retcode, err := jc.GetInt("BaseResponse.Ret")
    if err != nil {
        return nil, err
    }
    if retcode != 0 {
        return nil, fmt.Errorf("BaseResponse.Ret %d", retcode)
    }

    skl.List = skl.List[:0]
    skl1, _ := GetSyncKeyListFromJc(jc)
    skl.Count = skl1.Count
    skl.List = append(skl.List, skl1.List...)

    return body, nil
}
package main

import (
    "strconv"
    "time"
    "math/rand"
    "reflect"
)


func GetTimeStamp(len int) string {
    ts := strconv.FormatInt(time.Now().Unix(), len)
    return ts
}


func GetRandomStringFromNum(length int) string {
    bytes := []byte("0123456789")
    result := []byte{}
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    for i := 0; i < length; i++ {
        result = append(result, bytes[r.Intn(len(bytes))])
    }
    return string(result)
}


func GetUserInfoFromJc(jc *JsonConfig) (*User, error) {
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

func GetSyncKeyListFromJc(jc *JsonConfig) (*SyncKeyList, error) {
    is, err := jc.GetInterfaceSlice("SyncKey.List") //[]interface{}
    if err != nil {
        return nil, err
    }
    synks := make([]SyncKey, 0)
    for _, v := range is {
        // interface{}
        vm := v.(map[string]interface{})
        sk := SyncKey{
            Key: int(vm["Key"].(float64)),
            Val: int(vm["Val"].(float64)),
        }
        synks = append(synks, sk)
    }
    return &SyncKeyList{
        Count: len(synks),
        List:  synks,
    }, nil
}
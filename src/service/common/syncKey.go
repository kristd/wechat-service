package common

import (
    "strconv"
    "strings"
    "service/conf"
)

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

// s.String output synckey list in string
func (s *SyncKeyList) String() string {
    strs := make([]string, 0)

    for _, v := range s.List {
        strs = append(strs, strconv.Itoa(v.Key)+"_"+strconv.Itoa(v.Val))
    }
    return strings.Join(strs, "|")
}

func GetSyncKeyListFromJc(jc *conf.JsonConfig) (*SyncKeyList, error) {
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

func GetSessionGroupFromJc(jc *conf.JsonConfig) ([]*User, error) {

    return nil, nil
}

package conf

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"reflect"
	"strings"
)

type GlobalConfig struct {
	PORT        string `json:"port"`
	IMAGE_PATH  string `json:"image_path"`
	NOTIFY_ADDR string `json:"notify_addr"`
}

type JsonConfig struct {
	m  map[string]interface{}
	rb []byte
}

// XmlConfig: web api xml response struct
type XmlConfig struct {
	XMLName     xml.Name `xml:"error"`
	Ret         int      `xml:"ret"`
	Message     string   `xml:"message"`
	Skey        string   `xml:"skey"`
	Wxsid       string   `xml:"wxsid"`
	Wxuin       string   `xml:"wxuin"`
	PassTicket  string   `xml:"pass_ticket"`
	IsGrayscale int      `xml:"isgrayscale"`
}

var (
	Config *GlobalConfig
)

func LoadConfig(f string) bool {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		glog.Error("[LoadConfig] Read config file failed f:", f, " err:", err)
		return false
	}

	Config = &GlobalConfig{}
	err = json.Unmarshal(b, Config)
	if err != nil {
		glog.Error("[LoadConfig] Unmarshal config file failed err:", err, " content:", string(b))
		return false
	}

	glog.Infof(">>> [LoadConfig] Read config file [%v] ok!\n\n%#v\n\n", f, Config)
	return true
}

func LoadJsonConfigFromFile(path string) (*JsonConfig, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadJsonConfigFromBytes(b)
}

func LoadJsonConfigFromBytes(b []byte) (*JsonConfig, error) {
	var jm map[string]interface{}
	if err := json.Unmarshal(b, &jm); err != nil {
		return nil, err
	}
	s := &JsonConfig{
		m:  jm,
		rb: b,
	}
	return s, nil
}

func (s *JsonConfig) Dump() (string, error) {
	var rj bytes.Buffer
	if err := json.Indent(&rj, s.rb, "", "\t"); err != nil {
		return "", err
	}
	return string(rj.Bytes()), nil
}

// get("a.b.c")
func (s *JsonConfig) get(key string, m map[string]interface{}) (interface{}, error) {
	if m == nil {
		m = s.m
	}
	nodes := strings.Split(key, ".")
	for i := 0; i < len(nodes); i++ {
		if v, ok := m[nodes[i]]; ok {
			switch reflect.ValueOf(v).Kind() {
			case reflect.Map:
				if vv, okk := v.(map[string]interface{}); okk {
					// not end
					m = vv
				}
			default:
				return v, nil
			}
		} else {
			return nil, fmt.Errorf("[get] no value for key %s", key)
		}
	}
	return m, nil
}

// get("a.b.c")
func (s *JsonConfig) getSliceChilds(key string, m map[string]interface{}) ([]interface{}, error) {
	if m == nil {
		m = s.m
	}
	nodes := strings.Split(key, ".")
	for i := 0; i < len(nodes); i++ {
		if v, ok := m[nodes[i]]; ok {
			switch reflect.ValueOf(v).Kind() {
			case reflect.Map:
				if vv, okk := v.(map[string]interface{}); okk {
					// not slice
					m = vv
				}
			case reflect.Slice:
				result := make([]interface{}, 0)
				if vv, okk := v.([]interface{}); okk {
					// may not end
					for _, child := range vv {
						if v1, ok1 := child.(map[string]interface{}); ok1 {
							res, _ := s.get(strings.Join(nodes[i+1:], "."), v1)
							result = append(result, res)
						} else {
							result = append(result, child)
						}
					}
					return result, nil
				}
			}
		} else {
			return nil, fmt.Errorf("[getSliceChilds] no value for key %s", key)
		}
	}
	return nil, fmt.Errorf("[getSliceChilds] can't get []interface{}")
}

// user funcs
// leaf [{string}, {string}]
func (s *JsonConfig) GetSliceString(key string) ([]string, error) {
	is, err := s.getSliceChilds(key, nil)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	for _, v := range is {
		result = append(result, v.(string))
	}
	return result, nil
}

// leaf [{int}, {int}]
func (s *JsonConfig) GetSliceInt(key string) ([]int, error) {
	is, err := s.getSliceChilds(key, nil)
	if err != nil {
		return nil, err
	}
	result := make([]int, 0)
	for _, v := range is {
		result = append(result, int(v.(float64)))
	}
	return result, nil
}

// leaf [{int64}, {int64}]
func (s *JsonConfig) GetSliceInt64(key string) ([]int64, error) {
	is, err := s.getSliceChilds(key, nil)
	if err != nil {
		return nil, err
	}
	result := make([]int64, 0)
	for _, v := range is {
		result = append(result, int64(v.(float64)))
	}
	return result, nil
}

// child {}
func (s *JsonConfig) GetInterface(key string) (interface{}, error) {
	return s.get(key, nil)
}

// leaf [string, string, ...]
func (s *JsonConfig) GetStringSlice(key string) ([]string, error) {
	empty := []string{}
	f, err := s.get(key, nil)
	if err != nil {
		return empty, err
	}
	if _, ok := f.([]interface{}); !ok {
		return empty, fmt.Errorf("[GetStringSlice] value for key %s is not slice", key)
	}
	sf := f.([]interface{})
	ss := make([]string, len(sf))
	for i, v := range sf {
		if vv, ok := v.(string); ok {
			ss[i] = vv
		} else {
			return empty, fmt.Errorf("[GetStringSlice] %s[%d] is not a string", key, i)
		}
	}
	return ss, nil
}

// leaf string
func (s *JsonConfig) GetString(key string) (string, error) {
	f, err := s.get(key, nil)
	if err != nil {
		return "", err
	}
	if _, ok := f.(string); !ok {
		return "", fmt.Errorf("[GetString] value for key %s is not string", key)
	}
	return f.(string), nil
}

// leaf int
func (s *JsonConfig) GetInt(key string) (int, error) {
	f, err := s.get(key, nil)
	if err != nil {
		return 0, err
	}
	if _, ok := f.(float64); !ok {
		return 0, fmt.Errorf("[GetInt] value for key %s is not int", key)
	}
	return int(f.(float64)), nil
}

// leaf float
func (s *JsonConfig) GetFloat64(key string) (float64, error) {
	f, err := s.get(key, nil)
	if err != nil {
		return 0.0, err
	}
	if _, ok := f.(float64); !ok {
		return 0.0, fmt.Errorf("[GetFloat64] value for key %s is not float64", key)
	}
	return f.(float64), nil
}

// child [{}, {}, {}]
func (s *JsonConfig) GetInterfaceSlice(key string) ([]interface{}, error) {
	f, err := s.get(key, nil)
	if err != nil {
		return nil, err
	}
	if _, ok := f.([]interface{}); !ok {
		return nil, fmt.Errorf("[GetInterfaceSlice] value for key %s is not []interface{}", key)
	}
	return f.([]interface{}), nil
}

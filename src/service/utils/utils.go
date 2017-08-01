package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"service/conf"
	"strconv"
	"strings"
	"time"
	"github.com/golang/glog"
)

func GetTimeStamp(len int) string {
	ts := strconv.FormatInt(time.Now().Unix(), len)
	return ts
}

func GetRandomStringFromNum(rangeStr string, length int) string {
	bytes := []byte(rangeStr)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func LoadImage(url string) (fileName string, err error) {
	path := strings.Split(url, "/")
	if len(path) > 1 {
		fileName = path[len(path)-1]
	}

	fileName = conf.Config.IMAGE_PATH + fileName
	out, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	pix, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(out, bytes.NewReader(pix))
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func FilterNewMemberNickName(message string) string {
	newJoiner := ""
	for k, v := range conf.WELCOME_MESSAGE_PATTERN {
		r, err := regexp.Compile(v)
		if err != nil {
			return ""
		}

		newJoiner = r.FindString(message)
		if newJoiner != "" {
			switch k {
			case 0:
				newJoiner = strings.Replace(strings.Replace(newJoiner, "邀请", "", -1), "加入了群聊", "", -1)
			case 1:
				newJoiner = strings.Replace(newJoiner, "通过扫描", "", -1)
			case 2:
				newJoiner = strings.Replace(strings.Replace(newJoiner, "invited ", "", -1), " to the", "", -1)
			case 3:
				newJoiner = strings.Replace(newJoiner, " joined the group", "", -1)
			case 4:
				newJoiner = strings.Replace(newJoiner, "通过游戏中心加入群聊", "", -1)
			}
			break
		}
	}

	if newJoiner == "" {
		glog.Error("[FindNewJoinerName] New joiner name not match")
	}

	return newJoiner
}

func GetTimeNow() string {
	return time.Unix(time.Now().Unix(), 0).String()
}
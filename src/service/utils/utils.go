package utils

import (
	"bytes"
	"fmt"
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

func FindNewJoinerName(message string) (string, error) {
	newJoiner := ""
	for k, v := range conf.WELCOME_MESSAGE_PATTERN {
		r, err := regexp.Compile(v)
		if err != nil {
			return "", err
		}

		newJoiner = r.FindString(message)
		if len(newJoiner) != 0 {
			switch k {
			case 0:
				newJoiner = strings.Replace(strings.Replace(newJoiner, "邀请", "", -1), "加入了群聊", "", -1)
			case 1:
				newJoiner = strings.Replace(newJoiner, "通过扫描", "", -1)
			case 2:
				newJoiner = strings.Replace(strings.Replace(newJoiner, "invited ", "", -1), " to the", "", -1)
			case 3:
				newJoiner = strings.Replace(newJoiner, " joined the group", "", -1)
			}
			break
		}
	}

	fmt.Println(">>> newJoiner = ", newJoiner)

	if len(newJoiner) == 0 {
		return "", fmt.Errorf("[FindNewJoinerName] New joiner name empty")
	} else {
		return newJoiner, nil
	}
}

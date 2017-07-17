package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"service/conf"
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

	fileName = conf.IMG_PATH + fileName
	out, err := os.Create(fileName)

	defer out.Close()

	resp, err := http.Get(url)
	defer resp.Body.Close()

	pix, err := ioutil.ReadAll(resp.Body)
	_, err = io.Copy(out, bytes.NewReader(pix))

	return fileName, err
}

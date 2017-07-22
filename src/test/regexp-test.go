package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	//str1 := `"李晓云"邀请""""""""加入了群聊`
	str1 := "123444"
	str2 := `"李晓云通过扫描"通过扫描"TJ"分享的二维码加入了群聊`
	str3 := `TJ" invited "李晓云" to the group chat`
	str4 := `"李晓云" joined the group chat via your shared QR Code`

	r1 := regexp.MustCompile(`邀请(.+)加入了群聊`)
	r2 := regexp.MustCompile(`(.+)通过扫描`)
	r3 := regexp.MustCompile(`invited (.+) to the`)
	r4 := regexp.MustCompile(`(.+) joined the group`)

	fmt.Println("r1.result1 =", strings.Replace(strings.Replace(r1.FindString(str1), "邀请", "", -1), "加入了群聊", "", -1))
	fmt.Println("r2.result2 =", strings.Replace(r2.FindString(str2), "通过扫描", "", -1))

	fmt.Println("r3.result3 =", strings.Replace(strings.Replace(r3.FindString(str3), "invited ", "", -1), " to the", "", -1))
	fmt.Println("r4.result4 =", strings.Replace(r4.FindString(str4), " joined the group", "", -1))
}

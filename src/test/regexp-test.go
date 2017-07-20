package main

import (
	"fmt"
	"regexp"
)

func main() {
	str1 := "12\"34\"5"
	str2 := "12345"
	//regexp err will throw exception
	r1 := regexp.MustCompile(`\"(.*?)\"`)
	//regexp err will return error
	//r2, _ := regexp.Compile(`\"(.*?)\"`)

	fmt.Println("r1.result1 = ", r1.FindString(str1))
	fmt.Println("r2.result2 = ", r1.FindString(str2))
}

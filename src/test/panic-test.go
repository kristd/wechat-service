package main

import (
	"fmt"
)

func main() {
	defer func() {
		//catch exception
		if err := recover(); err != nil {
			fmt.Println("main err = ", err)
		}
		//defer stack
		fmt.Println("first defer")
	}()

	defer func() {
		//defer stack
		fmt.Println("second defer")
	}()
	f()
}

func f() {
	a := []string{"a", "b"}
	//throw exception
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("f err = ", err)
		}
	}()

	fmt.Println(a[3])
	//never reach
	fmt.Println("f end")
}

package main

import (
	"fmt"
	"os"
)

func Print(value interface{}) {
	fmt.Println(value)
}

func PrintErrorAndExit(err error) {
	if err != nil {
		println(err.Error())
		Exit()
	}
}

func Exit() {
	os.Exit(1)
}

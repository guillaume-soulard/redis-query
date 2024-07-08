package main

import "os"

func PrintErrorAndExit(err error) {
	if err != nil {
		println(err.Error())
		Exit()
	}
}

func Exit() {
	os.Exit(1)
}

package main

import (
	"os"
	"os/exec"
	"runtime"
)

var clearFuncMap = map[string]func(){}

func init() {
	clearFuncMap["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clearFuncMap["darwin"] = clearFuncMap["linux"]
	clearFuncMap["windows"] = func() {
		// TODO: 설명 보완
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func clearConsole() {
	if clearFunc, ok := clearFuncMap[runtime.GOOS]; ok {
		clearFunc()
	}
}

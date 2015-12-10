package gltext

import (
	"fmt"
	"runtime"
)

var IsDebug = false

func DebugPrefix() string {
	_, fn, line, _ := runtime.Caller(1)
	return fmt.Sprintf("DB: [%s:%d]", fn, line)
}

func TextDebug(message string) {
	if IsDebug {
		pc, fn, line, _ := runtime.Caller(1)
		fmt.Printf("[error] in %s[%s:%d] %s", runtime.FuncForPC(pc).Name(), fn, line, message)
	}
}

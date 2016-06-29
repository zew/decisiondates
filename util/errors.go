package util

import (
	"os"
	"strings"

	"github.com/zew/assessmentratedate/logx"
)

func CheckErr(err error) {
	defer logx.SL().Incr().Decr()
	if err != nil {
		logx.Printf("%v", err)
		str := strings.Join(logx.StackTrace(2, 3, 2), "\n")
		logx.Printf("\nStacktrace: \n%s", str)
		os.Exit(1)
	}
}

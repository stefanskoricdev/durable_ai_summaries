package util

import (
	"fmt"

	"github.com/fatih/color"
)

type CustomLogger struct{}

func (l CustomLogger) Debug(msg string, keyvals ...any) {}
func (l CustomLogger) Info(msg string, keyvals ...any)  {}
func (l CustomLogger) Warn(msg string, keyvals ...any)  {}
func (l CustomLogger) Error(msg string, keyvals ...any) {}

var infoStr = color.New(color.FgCyan, color.Bold).SprintfFunc()

func LogInfo(str string) {
	fmt.Println("")
	fmt.Println(infoStr(str))
	fmt.Println("")
}

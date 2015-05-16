package gslogger

import (
	"fmt"
	"strings"

	color "github.com/gsdocker/gslogger/console"
)

// Console .
func Console(msgfmt string, timefmt string) {
	console.timefmt = timefmt
	console.msgfmt = msgfmt
}

var console = newConsoleSink()

type consoleSink struct {
	timefmt string
	msgfmt  string
}

func newConsoleSink() *consoleSink {
	return &consoleSink{
		timefmt: "2006-01-02 15:04:05.999999",
		msgfmt:  "$ts ($file:$lines) [$tag] $source -- $content",
	}
}

func (sink *consoleSink) Recv(msg *Msg) {

	var tag string
	var f func(format string, a ...interface{})

	switch msg.Flag {
	case ASSERT:
		tag = "A"
		f = color.Red
	case ERROR:
		tag = "E"
		f = color.Red
	case WARN:
		tag = "W"
		f = color.Magenta
	case INFO:
		tag = "I"
		f = color.White
	case DEBUG:
		tag = "D"
		f = color.Cyan
	case VERBOSE:
		tag = "V"
		f = color.Blue
	default:
		tag = "U"
		f = color.Blue
	}

	content := sink.msgfmt

	content = strings.Replace(content, "$ts", msg.TS.Format(sink.timefmt), -1)
	content = strings.Replace(content, "$file", msg.File, -1)
	content = strings.Replace(content, "$lines", fmt.Sprintf("%02d", msg.Lines), -1)
	content = strings.Replace(content, "$tag", tag, -1)
	content = strings.Replace(content, "$source", msg.Log.String(), -1)
	content = strings.Replace(content, "$content", msg.Content, -1)

	f(content)
}

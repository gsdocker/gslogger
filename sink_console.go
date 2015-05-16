package gslogger

import "github.com/gsdocker/gsconfig"
import color "github.com/gsdocker/gslogger/console"

var console = newConsoleSink()

type consoleSink struct {
	format string
}

func newConsoleSink() *consoleSink {
	return &consoleSink{
		format: gsconfig.String("gslogger.timestamp", "2006-01-02 15:04:05.999999"),
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

	f(
		"%s (%s:%02d) [%s] %s -- %s",
		msg.TS.Format(sink.format),
		msg.File,
		msg.Lines,
		tag,
		msg.Log,
		msg.Content,
	)
}

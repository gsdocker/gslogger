package gslogger

import (
	"fmt"
	"time"

	"github.com/mgutz/ansi"
)

var console consoleSink

type consoleSink struct {
	format       string
	errorColor   func(string) string
	warnColor    func(string) string
	textColor    func(string) string
	debugColor   func(string) string
	fatalColor   func(string) string
	verboseColor func(string) string
}

func (sink *consoleSink) Recv(msg *Msg) {
	switch msg.Flag {
	case ERROR:
		sink.Error(msg.TS, msg.Log, msg.File, msg.Lines, msg.Content)
	case WARN:
		sink.Warn(msg.TS, msg.Log, msg.File, msg.Lines, msg.Content)
	case INFO:
		sink.Text(msg.TS, msg.Log, msg.File, msg.Lines, msg.Content)
	case DEBUG:
		sink.Debug(msg.TS, msg.Log, msg.File, msg.Lines, msg.Content)
	case ASSERT:
		sink.Fatal(msg.TS, msg.Log, msg.File, msg.Lines, msg.Content)
	case VERBOSE:
		sink.Verb(msg.TS, msg.Log, msg.File, msg.Lines, msg.Content)
	default:
		sink.Uknown(msg.TS, msg.Log, msg.File, msg.Lines, msg.Content)
	}
}

func (sink *consoleSink) Uknown(timestamp time.Time, log Log, file string, line int, content string) {

	fmt.Println(
		sink.verboseColor(
			fmt.Sprintf("%s (%s:%02d) [Uknown] %s -- %s",
				timestamp.Format(sink.format),
				file,
				line,
				log, content)))
}

func (sink *consoleSink) Verb(timestamp time.Time, log Log, file string, line int, content string) {

	fmt.Println(
		sink.verboseColor(
			fmt.Sprintf("%s (%s:%02d) [V] %s -- %s",
				timestamp.Format(sink.format),
				file,
				line,
				log, content)))
}

func (sink *consoleSink) Error(timestamp time.Time, log Log, file string, line int, content string) {

	fmt.Println(
		sink.errorColor(
			fmt.Sprintf("%s (%s:%02d) [E] %s -- %s",
				timestamp.Format(sink.format),
				file,
				line,
				log, content)))
}

func (sink *consoleSink) Warn(timestamp time.Time, log Log, file string, line int, content string) {

	fmt.Println(
		sink.warnColor(
			fmt.Sprintf("%s (%s:%02d) [W] %s -- %s",
				timestamp.Format(sink.format),
				file,
				line,
				log, content)))
}

func (sink *consoleSink) Text(timestamp time.Time, log Log, file string, line int, content string) {
	fmt.Println(
		sink.textColor(
			fmt.Sprintf("%s (%s:%02d) [T] %s -- %s",
				timestamp.Format(sink.format),
				file,
				line,
				log, content)))
}

func (sink *consoleSink) Debug(timestamp time.Time, log Log, file string, line int, content string) {
	fmt.Println(
		sink.debugColor(
			fmt.Sprintf("%s (%s:%02d) [D] %s -- %s",
				timestamp.Format(sink.format),
				file,
				line,
				log, content)))
}

func (sink *consoleSink) Fatal(timestamp time.Time, log Log, file string, line int, content string) {
	fmt.Println(
		sink.fatalColor(
			fmt.Sprintf("%s (%s:%02d) [A] %s -- %s",
				timestamp.Format(sink.format),
				file,
				line,
				log, content)))
}

func init() {
	console.fatalColor = ansi.ColorFunc("red+u")
	console.errorColor = ansi.ColorFunc("red")
	console.warnColor = ansi.ColorFunc("magenta")
	console.textColor = ansi.ColorFunc("white")
	console.debugColor = ansi.ColorFunc("cyan")
	console.verboseColor = ansi.ColorFunc("cyan+u")
	console.format = "2006-01-02 15:04:05.999999"
}

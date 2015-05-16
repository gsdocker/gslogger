package gslogger

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

//LEVEL 日志级别
type LEVEL int32

//内建日志级别
const (
	ASSERT LEVEL = (1 << iota)
	ERROR
	WARN
	INFO
	DEBUG
	VERBOSE
)

//Log 日志对象接口
type Log interface {
	//Flags get the log's flags
	Flags() LEVEL
	//NewFlags set the log level flags
	NewFlags(flags LEVEL)
	//V write verbose level log message
	V(format string, v ...interface{})
	//D write debug level log message
	D(format string, v ...interface{})
	//W write INFO level log message
	W(format string, v ...interface{})
	//I write INFO level log message
	I(format string, v ...interface{})
	//E write ERROR level log message
	E(format string, v ...interface{})
	//A write ASSERT level log message
	A(format string, v ...interface{})
	//NewSinks replace the log's backend list
	NewSink(sinks ...Sink)
	//Sink get log's sink list
	Sink() []Sink

	String() string
}

//Sink log message handle object
type Sink interface {
	//Recv 接收并处理日志消息
	Recv(msg *Msg)
}

//Msg 日志消息类型
type Msg struct {
	Flag    LEVEL     //log level of this msg
	TS      time.Time //timestamp of created time
	Log     Log       //log object which generate this msg
	File    string    //source code file name
	Lines   int       //the lines of source code file
	Content string    //log msg content
}

//fontend object which implement Log interface
type fontend struct {
	flags   LEVEL       //log level flags
	name    string      //log's name
	sinks   []Sink      //log's backend list
	service *LogService //log servie object
}

//LogService log service
type LogService struct {
	sync.Mutex                //log service’s locker
	Q          chan *Msg      //log message cached process queue
	Flags      LEVEL          //global scope log level flags
	Sinks      []Sink         //global backend list
	Loggers    map[string]Log //register log objects
	Exit       chan bool      //exit event
}

//stacktrace get the source attributes of log msg
func stacktrace(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)

	if !ok {
		file = "???"

		line = 0
	}

	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}

	return file, line
}

func (fd *fontend) write(flag LEVEL, format string, v ...interface{}) {

	file, line := stacktrace(3)

	msg := &Msg{
		Flag:    flag,
		TS:      time.Now(),
		Log:     fd,
		File:    file,
		Lines:   line,
		Content: fmt.Sprintf(format, v...),
	}

	fd.service.Dispatch(msg)

}

func (fd *fontend) String() string {
	return fd.name
}

//Flags get the log's flags
func (fd *fontend) Flags() LEVEL {
	return fd.flags
}

//NewFlags set log's level flags
func (fd *fontend) NewFlags(flags LEVEL) {
	fd.flags = flags
}

//V write verbose level log message
func (fd *fontend) V(format string, v ...interface{}) {
	if fd.flags&VERBOSE != 0 {
		fd.write(VERBOSE, format, v...)
	}
}

//D write debug level log message
func (fd *fontend) D(format string, v ...interface{}) {
	if fd.flags&DEBUG != 0 {
		fd.write(DEBUG, format, v...)
	}
}

//I write INFO level log message
func (fd *fontend) I(format string, v ...interface{}) {
	if fd.flags&INFO != 0 {
		fd.write(INFO, format, v...)
	}
}

//W write WARN level log message
func (fd *fontend) W(format string, v ...interface{}) {
	if fd.flags&WARN != 0 {
		fd.write(WARN, format, v...)
	}
}

//E write ERROR level log message
func (fd *fontend) E(format string, v ...interface{}) {
	if fd.flags&ERROR != 0 {
		fd.write(ERROR, format, v...)
	}
}

//A write ASSERT level log message
func (fd *fontend) A(format string, v ...interface{}) {
	if fd.flags&ASSERT != 0 {
		fd.write(ASSERT, format, v...)
	}
}

//NewSink replace the log's backend list
func (fd *fontend) NewSink(sinks ...Sink) {
	fd.sinks = sinks
}

//Sink replace the log's backend list
func (fd *fontend) Sink() []Sink {
	return fd.sinks
}

//NewService create new log service object
func NewService(cachesize int) *LogService {
	service := &LogService{
		Q:       make(chan *Msg, cachesize),
		Flags:   ASSERT | ERROR | INFO | DEBUG | WARN | VERBOSE,
		Loggers: make(map[string]Log),
		Exit:    make(chan bool, 1),
		Sinks:   []Sink{console},
	}

	go service.loop()
	return service
}

//Dispatch dispatch msg to backend handler
func (service *LogService) Dispatch(msg *Msg) {
	service.Q <- msg
}

//NewFlags set the global log level flags
func (service *LogService) NewFlags(flags LEVEL) {

	service.Lock()
	defer service.Unlock()
	service.Flags = flags
	for _, log := range service.Loggers {
		log.NewFlags(flags)
	}
}

//NewSink replace the log's backend list
func (service *LogService) NewSink(sinks ...Sink) {

	service.Lock()
	defer service.Unlock()
	service.Sinks = sinks
	for _, log := range service.Loggers {
		log.NewSink(sinks...)
	}
}

//Get get log object by name
func (service *LogService) Get(name string) Log {
	service.Lock()
	defer service.Unlock()
	if log, ok := service.Loggers[name]; ok {
		return log
	}

	log := &fontend{
		flags:   service.Flags,
		name:    name,
		sinks:   service.Sinks,
		service: service,
	}

	service.Loggers[name] = log

	return log
}

func (service *LogService) loop() {
	for msg := range service.Q {
		for _, sink := range msg.Log.Sink() {
			sink.Recv(msg)
		}
	}

	close(service.Exit)
}

//Join join until the log service process all cached msg
func (service *LogService) Join() {
	close(service.Q)
	for _ = range service.Exit {

	}
}

var global *LogService

func init() {
	global = NewService(56)
}

//NewFlags set the global log level flags
func NewFlags(flags LEVEL) {

	global.NewFlags(flags)
}

//NewSink replace the log's backend list
func NewSink(sinks ...Sink) {

	global.NewSink(sinks...)
}

//Get get log object by name
func Get(name string) Log {
	return global.Get(name)
}

//Join join until the log service process all cached msg
func Join() {
	global.Join()
}

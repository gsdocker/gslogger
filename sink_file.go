package gslogger

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//日志文件默认参数
const (
	DefaultCutSize = 20000000 //默认切片大小
	CompressDay    = 0        //压缩几天前的日志
)

//日志文件目录
var LogDir = os.Getenv("HOME") + "/gocode/log"

//Filelog 文件日志结构
type Filelog struct {
	flag        byte
	logfile     *os.File     //文件指针
	mutex       sync.RWMutex //读写锁
	date        string       //日志日期
	size        int64        //文件分片大小
	num         uint32       //分片计数器
	logname     string       //日志名字
	description string       //描述
	filename    string       //全路径文件名
	timeformat  string       //日志时间格式
}

//SetLogDir 设置日志目录
func SetLogDir(dir string) {
	LogDir = dir
}

//NewFilelog 建立文件日志对象
func NewFilelog(ln string, des string, cutsize int64) *Filelog {
	var fl Filelog
	fl.num = 0
	if cutsize == 0 {
		fl.size = DefaultCutSize
	} else {
		fl.size = cutsize
	}
	date := time.Now()
	fl.date = fmt.Sprintf("%04d%02d%02d", date.Year(), date.Month(), date.Day())
	fl.logname = ln
	fl.description = des
	fl.filename = fmt.Sprintf("%s/%s_%s_%s_%d.log", LogDir, fl.logname, fl.description, fl.date, fl.num)
	fl.logfile = nil
	fl.flag = 0xFF
	fl.timeformat = "2006-01-02 15:04:05.000000"
	return &fl
}

//Destory 删除日志文件对象
func (fl *Filelog) Destory() {
	if fl.logfile != nil {
		fl.logfile.Close()
		fl.logfile = nil
	}
	fl.flag = 0x00
}

//open 打开日志文件并判断大小及可用
func open(filename string, cutsize int64) *os.File {
	logfile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil
	}
	fi, err := logfile.Stat()
	if err != nil {
		logfile.Close()
		return nil
	}
	if fi.Size() > cutsize {
		logfile.Close()
		return nil
	}
	return logfile
}

//Getlogger 获取文件日志
func (fl *Filelog) Getlogger() {
	var cutsize int64
	if cutsize = fl.size; cutsize == 0 {
		cutsize = DefaultCutSize
	}
	if fl.logfile != nil {
		date := time.Now()
		temp := fmt.Sprintf("%04d%02d%02d", date.Year(), date.Month(), date.Day())

		if fl.date != temp {
			fl.logfile.Close()
			fl.logfile = nil
		} else if fi, err := fl.logfile.Stat(); err != nil {
			fl.logfile.Close()
			fl.logfile = nil
		} else if fi.Size() > cutsize {
			fl.logfile.Close()
			fl.logfile = nil
		}
	}
	if fl.logfile == nil {
		for {
			date := time.Now()
			temp := fmt.Sprintf("%04d%02d%02d", date.Year(), date.Month(), date.Day())
			if fl.date != temp {
				fl.date = temp
			}
			fl.filename = fmt.Sprintf("%s/%s_%s_%s_%d.log", LogDir, fl.logname, fl.description, fl.date, fl.num)
			fl.num++
			fl.logfile = open(fl.filename, cutsize)
			if fl.logfile != nil {
				break
			}
		}
	}
}

//Recv implement sink interface
func (fl *Filelog) Recv(msg *Msg) {
	var logrank string
	fl.mutex.Lock()
	defer fl.mutex.Unlock()
	if fl.flag != 0xFF {
		log.Fatal("the filelog is not been created")
	}
	fl.Getlogger()
	switch msg.Flag {
	case ASSERT:
		logrank = "A"
	case ERROR:
		logrank = "E"
	case WARN:
		logrank = "W"
	case INFO:
		logrank = "I"
	case DEBUG:
		logrank = "D"
	case VERBOSE:
		logrank = "V"
	default:
		logrank = "Uknown"
	}
	if fl.logfile != nil {
		fmt.Fprintf(fl.logfile, "%s (%s:%d):[%s] %s -- %s\n", msg.TS.Format(fl.timeformat), msg.File, msg.Lines, logrank, msg.Log, msg.Content)
	} else {
		fmt.Fprintf(os.Stdout, "%s (%s:%d):[%s] %s -- %s\n", msg.TS.Format(fl.timeformat), msg.File, msg.Lines, logrank, msg.Log, msg.Content)
	}
}

//MakeDir 初始化日志目录
func MakeDir() {
	var temp string
	temp = LogDir
	if err := os.Mkdir(temp, os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}
}

//CompressLog 压缩给定名字日志文件，现在仅将时间线以前的日志文件压到一起，不区分名字。可添加根据名字分别压缩的功能
func CompressLog(logname, description string) bool {

	dirname := LogDir

	date := time.Now()
	date = date.AddDate(0, 0, CompressDay)
	theday := fmt.Sprintf("%04d%02d%02d", date.Year(), date.Month(), date.Day())

	fw, err := os.OpenFile(dirname+"/"+theday+".tar.gz", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Println("create file failed.", err)
		return false
	}
	defer fw.Close()
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	dir, err := os.Open(dirname)
	if err != nil {
		log.Println(err)
		return false
	}
	defer dir.Close()
	fis, err := dir.Readdir(0)
	if err != nil {
		log.Println(err)
		return false
	}
	for _, fi := range fis {

		s := fi.Name()
		index := strings.IndexByte(s, '_')
		if index == -1 {
			continue
		}
		s = s[index+1:]
		index = strings.IndexByte(s, '_')
		if index == -1 {
			continue
		}
		s = s[index+1 : index+9]
		x, _ := strconv.Atoi(s)
		y, _ := strconv.Atoi(theday)
		if x > y {
			continue
		}
		fr, err := os.Open(dirname + "/" + fi.Name())
		if err != nil {
			log.Println("open log file failed.", err)
			return false
		}
		defer fr.Close()
		h := new(tar.Header)
		h.Name = fi.Name()
		h.Size = fi.Size()
		h.Mode = int64(fi.Mode())
		h.ModTime = fi.ModTime()
		err = tw.WriteHeader(h)
		if err != nil {
			log.Println("compress log file failed.", err)
			return false
		}
		_, err = io.Copy(tw, fr)
		if err != nil {
			log.Println("compress log file failed.", err)
			return false
		}
		tw.Flush()
		err = os.Remove(dirname + "/" + fi.Name())
		if err != nil {
			log.Println("remove log file failed", err)
		}
	}
	return true
}

//UncompressLog 解压缩指定日期日志 eg.theday:20140808
func UncompressLog(theday string) bool {
	fr, err := os.Open(LogDir + "/" + theday + ".tar.gz")
	if err != nil {
		log.Println("open tar file failed.", err)
		return false
	}
	defer fr.Close()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		log.Println(err)
		return false
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			return false
		}
		fw, err := os.OpenFile(LogDir+"/"+h.Name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return false
		}
		defer fw.Close()
		_, err = io.Copy(fw, tr)
		if err != nil {
			log.Println("uncompress log failed.", err)
			return false
		}
	}
	err = os.Remove(LogDir + "/" + theday + ".tar.gz")
	if err != nil {
		log.Println("remove file failed.", err)
		return false
	}
	return true
}

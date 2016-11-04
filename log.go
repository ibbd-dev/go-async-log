package asyncLog

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type asyncLogType struct {
	files map[string]*LogFile // 下标是文件名

	// 避免并发new对象
	sync.Mutex
}

type LogFile struct {
	filename string // 原始文件名（包含完整目录）
	flag     int    // 默认为log.LstdFlags

	// 同步设置
	sync struct {
		cycle     time.Duration // 同步数据到文件的周期，默认为1秒
		beginTime time.Time     // 开始同步的时间，判断同步的耗时
		status    syncStatus    // 同步状态
	}

	// 日志的等级
	level Priority

	// 缓存
	cache struct {
		use  bool     // 是否使用缓存
		data []string // 缓存数据
	}

	// 写log时的互斥锁
	writeMutex sync.Mutex

	// 文件切割
	logRotate struct {
		rotate LogRotate  // 默认按小时切割
		file   *os.File   // 文件操作对象
		suffix string     // 切割后的文件名后缀
		mutex  sync.Mutex // 文件名锁
	}
}

// log同步的状态
type syncStatus int

const (
	statusInit  syncStatus = iota // 初始状态
	statusDoing                   // 同步中
	statusDone                    // 同步已经完成
)

// 日志切割的方式
type LogRotate int

const (
	RotateHour LogRotate = iota // 按小时切割
	RotateDate                  // 按日期切割
)

const (
	// 写日志时前缀的时间格式
	logTimeFormat string = "2006-01-02 15:04:05"

	// 文件写入mode
	fileOpenMode = 0666

	// 文件Flag
	fileFlag = os.O_WRONLY | os.O_CREATE | os.O_APPEND

	// 换行符
	newlineChar = "\n"

	// 缓存切片的初始容量
	cacheInitCap = 64
)

// 是否需要Flag信息
const (
	NoFlag  = 0
	StdFlag = log.LstdFlags
)

// 异步日志变量
var asyncLog *asyncLogType

func init() {
	asyncLog = &asyncLogType{
		files: make(map[string]*LogFile),
	}

	timer := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				//now := time.Now()
				for _, file := range asyncLog.files {
					if file.sync.status != statusDoing {
						go file.flush()
					}
				}
			}
		}

	}()
}

func NewLogFile(filename string) *LogFile {
	asyncLog.Lock()
	if lf, ok := asyncLog.files[filename]; ok {
		asyncLog.Unlock()
		return lf
	}

	lf := &LogFile{
		filename: filename,
		flag:     StdFlag,
	}

	asyncLog.files[filename] = lf
	asyncLog.Unlock()

	// 默认按小时切割文件
	lf.logRotate.rotate = RotateHour

	// 默认开启缓存
	lf.cache.use = true

	// TODO 同步的时间周期，缓存开启才有效
	lf.sync.cycle = time.Second
	return lf
}

func (lf *LogFile) SetFlags(flag int) {
	lf.flag = flag
}

func (lf *LogFile) SetRotate(rotate LogRotate) {
	lf.logRotate.rotate = rotate
}

func (lf *LogFile) SetUseCache(useCache bool) {
	lf.cache.use = useCache
}

// 写缓存
func (lf *LogFile) Write(msg string) error {
	if lf.flag == StdFlag {
		msg = time.Now().Format(logTimeFormat) + " " + msg + newlineChar
	} else {
		msg = msg + newlineChar
	}

	if lf.cache.use {
		lf.writeMutex.Lock()
		lf.cache.data = append(lf.cache.data, msg)
		lf.writeMutex.Unlock()
		return nil
	}

	return lf.directWrite(msg)
}

//*********************** 以下是私有函数 ************************************

// 同步缓存到文件中
func (lf *LogFile) flush() error {
	lf.sync.status = statusDoing
	defer func() {
		lf.sync.status = statusDone
	}()

	// 写入log文件
	file, err := lf.openFileNoCache()
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 获取缓存数据
	lf.writeMutex.Lock()
	cache := lf.cache.data
	lf.cache.data = make([]string, 0, cacheInitCap)
	lf.writeMutex.Unlock()

	if len(cache) == 0 {
		return nil
	}

	_, err = file.WriteString(strings.Join(cache, ""))
	if err != nil {
		// 重试
		_, err = file.WriteString(strings.Join(cache, ""))
		if err != nil {
			panic(err)
		}
	}

	return nil
}

// 获取文件名的后缀
func (lf *LogFile) getFilenameSuffix() string {
	if lf.logRotate.rotate == RotateDate {
		return time.Now().Format("20060102")
	}
	return time.Now().Format("2006010215")
}

// 直接写入日志文件
func (lf *LogFile) directWrite(msg string) error {
	file, err := lf.openFile()
	//file, err := lf.openFileNoCache()
	if err != nil {
		panic(err)
	}
	//defer file.Close()

	lf.writeMutex.Lock()
	_, err = file.WriteString(msg)
	lf.writeMutex.Unlock()

	return err
}

// 打开日志文件
func (lf *LogFile) openFile() (*os.File, error) {
	suffix := lf.getFilenameSuffix()

	lf.logRotate.mutex.Lock()
	defer lf.logRotate.mutex.Unlock()

	if suffix == lf.logRotate.suffix {
		return lf.logRotate.file, nil
	}

	logFilename := lf.filename + "." + suffix
	file, err := os.OpenFile(logFilename, fileFlag, fileOpenMode)
	if err != nil {
		// 重试
		file, err = os.OpenFile(logFilename, fileFlag, fileOpenMode)
		if err != nil {
			return file, err
		}
	}

	// 关闭旧的文件
	if lf.logRotate.file != nil {
		lf.logRotate.file.Close()
	}

	lf.logRotate.file = file
	lf.logRotate.suffix = suffix
	return file, nil
}

// 打开日志文件(不缓存句柄)
func (lf *LogFile) openFileNoCache() (*os.File, error) {
	logFilename := lf.filename + "." + lf.getFilenameSuffix()
	file, err := os.OpenFile(logFilename, fileFlag, fileOpenMode)
	if err != nil {
		// 重试
		file, err = os.OpenFile(logFilename, fileFlag, fileOpenMode)
		if err != nil {
			return file, err
		}
	}

	return file, nil
}

# 异步日志库

## Install

```sh
go get -u github.com/ibbd-dev/go-async-log
```

## 实现的功能及说明

- 多个日志文件写入：例如错误信息一个文件，测试信息一个文件等
- 日志自动切割：前期支持按小时或者天切割
- 支持批量写入：最小单位为秒，不同的文件可以设置不同的写入频率（周期性写入，程序挂掉的时候最多可能会丢一个周期的数据，重要数据不能采用该方式
- 同时支持实时写入文件，使用文件系统缓存（只要系统不挂，就不会有问题）
- 错误等级实现
- 可以写入json数据
- 时间格式采用`RFC3339`，格式如`2006-01-02T15:04:05Z07:00`

## 配置项

- 文件名
- 日志记录的等级
- 自动切割周期：默认按小时
- 批量写入周期：默认每秒写入一次
- 异常等级：
- 是否需要Flags：默认需要

## Example

普通写入日志文件

```go
lf := asyncLog.NewLogFile("/tmp/test.log")

// 设置按天切割文件，如果默认则是按小时
lf.SetRotate(asyncLog.RotateDate)

_ = lf.Write("lf: hello world")

// 注意：因为是每秒写入一次，所以这里需要暂停一下
time.Sleep(time.Second * 2)

```

写入错误等级文件

```go
infoFile := asyncLog.NewLevelLog("/tmp/test-info.log", asyncLog.LevelInfo)  // 只有Info级别或者以上级别的日志才会被记录
infoFile.Debug("hello world") // 该日志不会写入文件
infoFile.Info("hello world")
infoFile.Error("hello world")

// 需要改变日志写入等级时，例如测试阶段
infoFile.SetLevel(asyncLog.LevelDebug)

time.Sleep(time.Second * 2)
```

## 性能数据

不缓存内容的时候，如果不对文件句柄进行缓存重用，性能是比较低的，如下：（这是旧版本）

```sh
# go test -bench=".*"
BenchmarkWrite-4          	 3000000	       444 ns/op
BenchmarkWriteNoCache-4   	  300000	      4400 ns/op
```

对句柄进行缓存重用之后，性能如下：

```sh
BenchmarkWrite-4          	 3000000	       570 ns/op
BenchmarkWriteNoCache-4   	 1000000	      2304 ns/op
```

结论：对句柄进行缓存，是能大大提升效率的。

-------

性能比`github.com/ibbd-dev/go-tools/logfile`至少提升一个数量级


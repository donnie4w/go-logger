## go-logger 是 go的灵活高效日志管理库

------------

### 功能特点

1. **日志级别设置**：支持动态调整日志级别，以便在不同环境下控制日志的详细程度。
2. **格式化输出**：支持自定义日志的输出格式，包括时间戳、日志级别、日志位置 等元素。
3. **文件数回滚**：支持按照日志文件数自动文件回滚，并防止日志文件数过多。
4. **文件压缩**：支持压缩归档日志文件。
5. **支持标准库log/slog日志文件管理**：支持标准库文件切割，压缩等功能。
6. **外部处理函数**：支持自定义外部处理函数。
7. **日志堆栈信息**：支持日志记录点可以回溯到程序入口点的所有函数调用序列，包括每一步函数调用的文件名，函数名，行号
8. **日志级别独立日志格式输出**：支持不同日志级别 指定不同的日志输出格式。

### `go-logger` +  `slog`

-  支持 直接作为go 标准库  `log/slog`  的日志文件管理器，实现 `slog`的日志文件按小时，天，月份，文件大小等多种方式进行日志文件切割，同时也支持按文件大小切分日志文件后,压缩归档日志文件。
- `go-logger` + slog 内存分配与性能 与 slog直接写日志文件一致。
- 详细参见[使用文档](https://tlnet.top/logdoc "使用文档")

### [使用文档](https://tlnet.top/logdoc "使用文档")

------------

### 一. 设置日志打印格式 `SetFormat`

##### 如： SetFormat(FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME)

##### 默认格式:  `FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME`

	不格式化，只打印日志内容		FORMAT_NANO		无格式
	长文件名及行数			FORMAT_LONGFILENAME	全路径
	短文件名及行数			FORMAT_SHORTFILENAME	如：logging_test.go:10
	精确到日期			FORMAT_DATE		如：2023/02/14
	精确到秒			        FORMAT_TIME		如：01:33:27
	精确到微秒			FORMAT_MICROSECONDS	如：01:33:27.123456
    日志级别标识                     FORMAT_LEVELFLAG        如：[Debug],[Info],[Warn][Error][Fatal]             
    调用函数                         FORMAT_FUNC             调用函数的函数名，若设置，则出现在文件名之后


#### 示例：

```go
logger.SetFormat(logger.FORMAT_LEVELFLAG | logger.FORMAT_LONGFILENAME | logger.FORMAT_TIME)
logger.Error("错误信息：文件未找到")
// 输出:
// [ERROR]/usr/log/logging/main.go:20 10:45:00: 错误信息：文件未找到
```

### 二. 设置日志标识输出格式  `SetFormatter`

######  `SetFormatter("{level} {time} {file}:{message}\n")`

##### 默认格式：`"{level} {time} {file}:{message}\n"`

	{level}        日志级别信息：如[Debug],[Info],[Warn],[Error],[Fatal]
	{time}         日志时间信息
	{file}         文件位置行号信息
	{message}      日志内容

##### 说明：除了关键标识  `{message}`  `{time}`  `{file}`  `{level}` 外，其他内容原样输出，如 | ， 空格，换行  等

###### 通过修改 `formatter`，可以自由定义输出格式，例如：

```go
logger.SetFormatter("{time} - {level} - {file} - {message}\n")
logger.Info("日志初始化完成")
// 输出:
// 2023/08/09 10:30:00 - [INFO] - main.go:12 - 日志初始化完成
```

------------

### 三. 日志级别 `SetLevel`  `SetLevelOption`

#####  DEBUG < INFO < WARN < ERROR < FATAL

###### 关闭所有日志 `SetLevel(OFF)`

`go-logger` 支持多种日志级别，从 `DEBUG` 到 `FATAL`，并可以通过 `SetLevel` 方法设置日志的输出级别：

```go
logger.SetLevel(logger.INFO)
logger.Debug("调试信息：这条日志不会被打印")
logger.Info("信息：这条日志会被打印")
```

##### 此外，可以通过 `SetLevelOption` 为不同的日志级别设置独立的日志输出格式：

```go
logger.SetLevelOption(logger.LEVEL_DEBUG, &logger.LevelOption{
    Format: logger.FORMAT_SHORTFILENAME | logger.FORMAT_TIME,
})
logger.SetLevelOption(logger.LEVEL_WARN, &logger.LevelOption{
    Format: logger.FORMAT_LONGFILENAME | logger.FORMAT_DATE | logger.FORMAT_FUNC,
})
```

##### 示例:分别给不同的日志级别设置不一样的输出格式
```go
func TestLevelOptions(t *testing.T) {
  SetLevelOption(LEVEL_DEBUG, &LevelOption{Format: FORMAT_LEVELFLAG | FORMAT_TIME | FORMAT_SHORTFILENAME})
  SetLevelOption(LEVEL_INFO, &LevelOption{Format: FORMAT_LEVELFLAG})
  SetLevelOption(LEVEL_WARN, &LevelOption{Format: FORMAT_LEVELFLAG | FORMAT_TIME | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_FUNC})
  
  Debug("this is a debug message")
  Info("this is a info message")
  Warn("this is a warn message")
}
```
###### 执行结果
```go
[DEBUG]18:53:55 logging_test.go:176 this is a debug message
[INFO]this is a info message
[WARN]2024/08/07 18:53:55 logging_test.go:TestLevelOptions:178 this is a warn message
```

### 四. 文件日志

##### go-logger支持日志信息写入文件，并提供文件分割的多种策略与压缩备份等特性

###### 使用全局log对象时，直接调用设置方法：
```go
SetRollingDaily()		按日期分割
SetRollingByTime()		可按 小时，天，月 分割日志
SetRollingFile()		指定文件大小分割日志
SetRollingFileLoop()		指定文件大小分割日志，并指定保留最大日志文件数
SetGzipOn(true)			压缩分割的日志文件 
```
#### 多实例：
```go
log1 := NewLogger()
log1.SetRollingDaily("", "logMonitor.log")

log12:= NewLogger()
log2.SetRollingDaily("", "logBusiness.log")
```
#### 1. 按日期分割日志文件
```go
log.SetRollingDaily("/var/logs", "log.txt")
// 每天按 log_20221015.txt格式 分割
//若 log_20221015.txt已经存在，则生成 log_20221015.1.txt ，log_20221015.2.txt等文件

log.SetRollingByTime("/var/logs", "log.txt",MODE_MONTH)
//按月份分割日志，跨月时，保留上月份日志，如：
//    log_202210.txt
//    log_202211.txt
//    log_202212.txt

log.SetRollingByTime("/var/logs", "log.txt",MODE_HOUR)
//按小时分割日志, 如：
//    log_2022101506.txt
//    log_2022101507.txt
//    log_2022101508.txt
```
#### 2. 按文件大小分割日志文件
```go
log.SetRollingFile("/var/logs", "log.txt", 300, MB)
//当文件超过300MB时，按log.1.txt，log.2.txt 格式备份
//目录参数可以为空，则默认当前目录。

log.SetRollingFileLoop("/var/logs", "log.txt", 300, MB, 50)
//设置日志文件大小最大为300M
//日志文件只保留最新的50个
```

------

### 五. Option参数 `SetOption`

###### 通过 `Option` 参数可以更加灵活地配置日志。`Option` 包含多个配置项，使得日志配置更加清晰和易于维护。

```go
logger.SetOption(&logger.Option{
  Level:       logger.LEVEL_DEBUG,
  Console:     false,
  Format:      logger.FORMAT_LEVELFLAG | logger.FORMAT_SHORTFILENAME | logger.FORMAT_DATE | logger.FORMAT_TIME,
  Formatter:   "{level} {time} {file}:{message}\n",
  FileOption:  &logger.FileSizeMode{
    Filename:  "app.log",
    Maxsize:   1 << 30,  // 1GB
    Maxbuckup: 10,
    IsCompress: true,
  },
})
```

##### 属性说明：
```text
Level           ：日志级别
Console         ：控制台打印
Format          ：日志格式，默认：FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
Formatter       ：日志输出  默认："{level}{time} {file}:{mesaage}\n"
FileOption      ：日志文件接口参数
Stacktrace      ：开启日志堆栈信息记录的日志级别
CustomHandler   ：自定义日志处理函数，返回true时，继续执行打印程序，返回false时，不再执行打印程序
```
1. #### FileOption介绍

  - ###### FileOption为接口，有FileSizeMode与FileTimeMode两个实现对象
  - ###### `FileTimeMode` 按时间滚动备份日志文件
    ```text
    Filename   日志文件路径
    Timemode   按小时，天，月份：MODE_HOUR，MODE_DAY，MODE_MONTH
    Maxbuckup  最多备份日志文件数
    IsCompress  备份文件是否压缩
    ```

  - ###### `FileSizeMode` 按文件大小滚动备份日志文件
    ```text
	Filename   日志文件路径
	Maxsize    日志文件大小的最大值，超过则滚动备份
	Maxbuckup  最多备份日志文件数
	IsCompress  备份文件是否压缩
    ```

  - ##### SetOption 示例1
    ```go
    // debug级别，关闭控制台日志打印，按天备份日志，最多日志文件数位10，备份时压缩文件，日志文件名为 testlogtime.log
    SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileTimeMode{Filename: "testlogtime.log", Maxbuckup: 10, IsCompress: true, Timemode: MODE_DAY}})
    ```
  - ##### SetOption 示例2
    ```go
    // debug级别，关闭控制台日志打印，按文件大小备份日志，按每文件大小为1G时备份一个文件，  最多日志文件数位10，备份时压缩文件，日志文件名为 testlog.log
    SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileSizeMode{Filename: "testlog.log", Maxsize: 1<<30, Maxbuckup: 10, IsCompress: true}})
    ```

2.  ##### **Stacktrace** 堆栈日志

  - ###### 栈追踪日志功能可以记录日志记录点到程序入口点的所有函数调用序列，包括每一步函数调用的文件名、函数名和行号。这对于调试和错误分析非常有帮助。
  - 当日志级别为 `WARN` 或更高时，日志记录将包含完整的调用栈信息。
  - **示例**

    ```go
    func TestStacktrace(t *testing.T) {
        SetOption(&Option{Console: true, Stacktrace: LEVEL_WARN, Format: FORMAT_LEVELFLAG | FORMAT_DATE | FORMAT_TIME | FORMAT_SHORTFILENAME | FORMAT_FUNC})
        Debug("this is a debug message")
        Stacktrace1()
    }
    
    func Stacktrace1() {
        Info("this is a info message")
        Stacktrace2()
    }
    
    func Stacktrace2() {
        Warn("this is a warn message")
        Stacktrace3()
    }
    
    func Stacktrace3() {
        Error("this is a error message")
        Fatal("this is a fatal message")
    }
    ```

    ##### 执行结果
    ```go
    [DEBUG]2024/08/07 20:22:40 logging_test.go:TestStacktrace:151 this is a debug message
    [INFO]2024/08/07 20:22:40 logging_test.go:Stacktrace1:156 this is a info message
    [WARN]2024/08/07 20:22:40 logging_test.go:Stacktrace2:161#logging_test.go:Stacktrace1:157#logging_test.go:TestStacktrace:152#testing.go:tRunner:1689#asm_amd64.s:goexit:1695 this is a warn message
    [ERROR]2024/08/07 20:22:40 logging_test.go:Stacktrace3:166#logging_test.go:Stacktrace2:162#logging_test.go:Stacktrace1:157#logging_test.go:TestStacktrace:152#testing.go:tRunner:1689#asm_amd64.s:goexit:1695 this is a error message
    [FATAL]2024/08/07 20:22:40 logging_test.go:Stacktrace3:167#logging_test.go:Stacktrace2:162#logging_test.go:Stacktrace1:157#logging_test.go:TestStacktrace:152#testing.go:tRunner:1689#asm_amd64.s:goexit:1695 this is a fatal message    ```
    ```
    
3. #### `CustomHandler` 自定义函数，可以自定义处理逻辑
- **示例**
  ```go
  func TestCustomHandler(t *testing.T) {
      SetOption(&Option{Console: true, CustomHandler: func(lc *LogContext) bool {
          fmt.Println("level:", levelname(lc.Level))
          fmt.Println("message:", fmt.Sprint(lc.Args...))
          if lc.Level == LEVEL_ERROR {
              return false //if error mesaage , do not print
          }
          return true
      },
      })
      Debug("this is a debug message")
      Info("this is a info message")
      Warn("this is a warn message")
      Error("this is a error message")
  }
  ```
- ##### 执行结果：根据CustomHandler的逻辑，不打印Error日志
  ```go
  level: debug
  message: this is a debug message
  [DEBUG]2024/08/07 18:51:56 logging_test.go:126 this is a debug message
  level: info
  message: this is a info message
  [INFO]2024/08/07 18:51:56 logging_test.go:127 this is a info message
  level: warn
  message: this is a warn message
  [WARN]2024/08/07 18:51:56 logging_test.go:128 this is a warn message
  level: error
  message: this is a error message
  ```

------------

### 六. 控制台日志设置 `SetConsole`
```go
//全局log：
SetConsole(false)  //控制台不打日志,默认值true
//实例log：
log.SetConsole(false)  //控制台不打日志,默认值true
```
***

### 七. 校正打印时间  `TIME_DEVIATION`
###### 有时在分布式环境中，可能存在不同机器时间不一致的问题，`go-logger` 允许通过 `TIME_DEVIATION` 参数来进行时间校正。

```go
logger.TIME_DEVIATION = 1000 // 将日志时间校正 +1微妙
```

------

## 性能测试：

```go
cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz
BenchmarkSerialZap
BenchmarkSerialZap-4                      714796              5469 ns/op             336 B/op          6 allocs/op
BenchmarkSerialZap-8                      675508              5316 ns/op             337 B/op          6 allocs/op
BenchmarkSerialLogger
BenchmarkSerialLogger-4                   749774              4458 ns/op             152 B/op          4 allocs/op
BenchmarkSerialLogger-8                   793208              4321 ns/op             152 B/op          4 allocs/op
BenchmarkSerialLoggerNoFORMAT
BenchmarkSerialLoggerNoFORMAT-4           977128              3767 ns/op             128 B/op          2 allocs/op
BenchmarkSerialLoggerNoFORMAT-8          1000000              3669 ns/op             128 B/op          2 allocs/op
BenchmarkSerialLoggerWrite
BenchmarkSerialLoggerWrite-4              856617              3659 ns/op             112 B/op          1 allocs/op
BenchmarkSerialLoggerWrite-8             1000000              3576 ns/op             112 B/op          1 allocs/op
BenchmarkSerialNativeGoLog
BenchmarkSerialNativeGoLog-4              892172              4488 ns/op             232 B/op          2 allocs/op
BenchmarkSerialNativeGoLog-8              798291              4327 ns/op             232 B/op          2 allocs/op
BenchmarkSerialSlog
BenchmarkSerialSlog-4                     634228              5602 ns/op             328 B/op          6 allocs/op
BenchmarkSerialSlog-8                     646191              5481 ns/op             328 B/op          6 allocs/op
BenchmarkSerialSlogAndLogger
BenchmarkSerialSlogAndLogger-4            626898              5671 ns/op             328 B/op          6 allocs/op
BenchmarkSerialSlogAndLogger-8            657820              5622 ns/op             328 B/op          6 allocs/op
BenchmarkParallelZap
BenchmarkParallelZap-4                    430472              7818 ns/op             336 B/op          6 allocs/op
BenchmarkParallelZap-8                    449402              7771 ns/op             337 B/op          6 allocs/op
BenchmarkParallelLogger
BenchmarkParallelLogger-4                 639826              5398 ns/op             152 B/op          4 allocs/op
BenchmarkParallelLogger-8                 604308              5532 ns/op             152 B/op          4 allocs/op
BenchmarkParallelLoggerNoFORMAT
BenchmarkParallelLoggerNoFORMAT-4         806749              4311 ns/op             128 B/op          2 allocs/op
BenchmarkParallelLoggerNoFORMAT-8         790284              4592 ns/op             128 B/op          2 allocs/op
BenchmarkParallelLoggerWrite
BenchmarkParallelLoggerWrite-4            764610              4141 ns/op             112 B/op          1 allocs/op
BenchmarkParallelLoggerWrite-8            880222              4079 ns/op             112 B/op          1 allocs/op
BenchmarkParallelNativeGoLog
BenchmarkParallelNativeGoLog-4            609134              5652 ns/op             232 B/op          2 allocs/op
BenchmarkParallelNativeGoLog-8            588201              5806 ns/op             232 B/op          2 allocs/op
BenchmarkParallelSLog
BenchmarkParallelSLog-4                   620878              5624 ns/op             328 B/op          6 allocs/op
BenchmarkParallelSLog-8                   636448              5532 ns/op             328 B/op          6 allocs/op
BenchmarkParallelSLogAndgoLogger
BenchmarkParallelSLogAndgoLogger-4        612314              5612 ns/op             328 B/op          6 allocs/op
BenchmarkParallelSLogAndgoLogger-8        633426              5596 ns/op             328 B/op          6 allocs/op
```

#### 压测结果分析

**日志记录库和方法：**
1. **Zap**：这是一个uber开发的高性能日志库。
2. **Logger**：go-logger日志库。
3. **Native Go Log**： Go 内置的 log 包。
4. **Slog**：这是 Go 1.19 引入的新标准日志库。
5. **Slog 和 go-logger 结合**：指同时使用go-logger作为slog的日志文件管理库。


##### 1. 基准测试指标解释：

*    **-4 和 -8**: 这些数字表示运行基准测试时使用的 CPU 核心数。-4 表示使用 4 个核心，而 -8 表示使用 8 个核心。
*    **ns/op**: 每次日志记录操作所需的平均时间（以纳秒为单位）。
*    **B/op**: 每次日志记录操作分配的平均内存大小（以字节为单位）。
*    **allocs/op**: 每次日志记录操作产生的分配次数。

##### 2. 串行日志记录结果：

*    **Zap**: 在 4 核心上有 5469 ns/op 的性能，在 8 核心上有 5316 ns/op 的性能。
*    **go-logger**: 在 4 核心上有 4458 ns/op 的性能，在 8 核心上有 4321 ns/op 的性能。
*    **go-logger(无格式)**: 在 4 核心上有 3767 ns/op 的性能，在 8 核心上有 3669 ns/op 的性能。
*    **go-logger(写操作)**: 在 4 核心上有 3659 ns/op 的性能，在 8 核心上有 3576 ns/op 的性能。
*    **Native Go Log**: 在 4 核心上有 4488 ns/op 的性能，在 8 核心上有 4327 ns/op 的性能。
*    **Slog**: 在 4 核心上有 5602 ns/op 的性能，在 8 核心上有 5481 ns/op 的性能。
*    **Slog 和 go-logger** 结合: 在 4 核心上有 5671 ns/op 的性能，在 8 核心上有 5622 ns/op 的性能。

##### 3. 并行日志记录结果：

*    **Zap**: 在 4 核心上有 7818 ns/op 的性能，在 8 核心上有 7771 ns/op 的性能。
*    **go-logger**: 在 4 核心上有 5398 ns/op 的性能，在 8 核心上有 5532 ns/op 的性能。
*    **go-logger (无格式)**: 在 4 核心上有 4311 ns/op 的性能，在 8 核心上有 4592 ns/op 的性能。
*    **go-logger (写操作)**: 在 4 核心上有 4141 ns/op 的性能，在 8 核心上有 4079 ns/op 的性能。
*    **Native Go Log**: 在 4 核心上有 5652 ns/op 的性能，在 8 核心上有 5806 ns/op 的性能。
*    **Slog**: 在 4 核心上有 5624 ns/op 的性能，在 8 核心上有 5532 ns/op 的性能。
*    **Slog 和go-logger 结合**: 在 4 核心上有 5612 ns/op 的性能，在 8 核心上有 5596 ns/op 的性能。

##### 4. 结果分析：

*    **Zap** 在串行模式下提供了较好的性能，但在并行模式下的性能有所下降。
*    **go-logger(写操作)** 在串行和并行模式下均表现出了最佳性能。
*    **go-logger(无格式)** 通过移除格式化步骤显著提高了性能。
*    **Native Go Log** 在串行和并行模式下性能接近于 go-logger。
*    **Slog** 的性能与 **Zap** 和 **go-logger** 相比略逊一筹。
*    **Slog** 和 **go-logger** 结合 的性能与 **Slog** 相近

##### 5. 结论

*    **从压测结果可以看到，在相同格式下，无论是串行还是高并发场景中，go-logger均表现出最佳性能和最小的内存分配。**
*    **内置库Log的性能 接近go-logger, 但它可能没有提供同样的灵活性.**
*    **go-logger作为slog日志文件管理库，无论内存分配还是性能，都与单独使用slog的效果相同，不会引入额外的性能开销。**
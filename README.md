## go-logger 是go 高性能日志库

------------

### 功能特点

- **日志级别设置**：允许动态调整日志级别，以便在不同环境下控制日志的详细程度。
- **格式化输出**：支持自定义日志的输出格式，包括时间戳、日志级别、日志位置 等元素。
- **文件数回滚**：支持按照日志文件数自动文件回滚，并防止日志文件数过多。
- **文件压缩**：支持压缩归档日志文件。
- **支持标准库log/slog日志文件管理**：支持标准库文件切割，压缩等功能。
- **外部处理函数**：支持自定义外部处理函数。
- **日志堆栈信息**：日志记录点可以回溯到程序入口点的所有函数调用序列，包括每一步函数调用的文件名，函数名，行号
- **日志级别独立日志格式输出**：支持不同日志级别 指定不同的日志输出格式。

### go-logger +  slog

-  支持 直接作为go 标准库  log/slog  的日志文件管理器，实现 slog的日志文件按小时，天，月份，文件大小等多种方式进行日志文件切割，同时也支持按文件大小切分日志文件后,压缩归档日志文件。
- go-logger + slog 内存分配与性能 与 slog直接写日志文件一致。
- 详细参见[使用文档](https://tlnet.top/logdoc "使用文档")

#### [使用文档](https://tlnet.top/logdoc "使用文档")

------------

#### 日志级别打印：

###### 调用 Debug()，Info()，Warn(), Error() ,Fatal() 级别由低到高

### 设置日志打印格式：

##### 如： SetFormat(FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME)

###### FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME 为默认格式

##### 不调用SetFormat()时，使用默认格式

	不格式化，只打印日志内容		FORMAT_NANO		无格式
	长文件名及行数			FORMAT_LONGFILENAME	全路径
	短文件名及行数			FORMAT_SHORTFILENAME	如：logging_test.go:10
	精确到日期			FORMAT_DATE		如：2023/02/14
	精确到秒			        FORMAT_TIME		如：01:33:27
	精确到微秒			FORMAT_MICROSECNDS	如：01:33:27.123456
    日志级别标识                     FORMAT_LEVELFLAG        如：[Debug],[Info],[Warn][Error][Fatal]             
    调用函数                         FORMAT_FUNC             调用函数的函数名，若设置，则出现在文件名之后
##### 打印结果形如：

###### [DEBUG]2023/02/14 01:33:27 logging_test.go 10: 11111111111111

### 设置日志标识输出格式  `SetFormatter`

######  SetFormatter("{level} {time} {file}:{message}\n")

###### 默认格式："{level} {time} {file}:{message}\n"

	{level}        日志级别信息：如[Debug],[Info],[Warn],[Error],[Fatal]
	{time}         日志时间信息
	{file}         文件位置行号信息
	{message}      日志内容

###### 说明：除了关键标识 {message}  {time}  {file}  {level} 外，其他内容原样输出，如 | ， 空格，换行  等

------------

### 日志级别

#####  DEBUG < INFO < WARN < ERROR < FATAL

###### 关闭所有日志 SetLevel(OFF)

#### 说明：

	若设置 INFO
	如：SetLevel(INFO)
	则 所有 Debug("*********")   不再打印出来
	所以调试阶段，常设置为默认级别ALL，或DEBUG，打印出项目中所有日志，包括调试日志
	若设置 OFF
	SetLevel(OFF)
	则 所有日志不再打印出来
	所以正式环境，常设置为ERROR或以上的日志级别，项目中Debug()，Info(),warn()等日志不再打印出来，具体视实际需求设置


#### 需将日志写入文件时，则要设置日志文件

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
log.SetRollingDaily("d:/foldTest", "log.txt")
// 每天按 log_20221015.txt格式 分割
//若 log_20221015.txt已经存在，则生成 log_20221015.1.txt ，log_20221015.2.txt等文件

log.SetRollingByTime("d:/foldTest", "log.txt",MODE_MONTH)
//按月份分割日志，跨月时，保留上月份日志，如：
//    log_202210.txt
//    log_202211.txt
//    log_202212.txt

log.SetRollingByTime("d:/foldTest", "log.txt",MODE_HOUR)
//按小时分割日志, 如：
//    log_2022101506.txt
//    log_2022101507.txt
//    log_2022101508.txt
```
#### 2. 按文件大小分割日志文件
```go
log.SetRollingFile("d:/foldTest", "log.txt", 300, MB)
//当文件超过300MB时，按log.1.txt，log.2.txt 格式备份
//目录参数可以为空，则默认当前目录。

log.SetRollingFileLoop(`d:/foldTest`, "log.txt", 300, MB, 50) 
//设置日志文件大小最大为300M
//日志文件只保留最新的50个
```

------

### Option参数

###### 建议通过option参数设置log的所有参数，更加易于维护
```text
Level           ：日志级别
Console         ：控制台打印
Format          ：日志格式，默认：FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
Formatter       ：日志输出  默认："{level}{time} {file}:{mesaage}\n"
FileOption      ：日志文件接口参数
Stacktrace      ：开启日志堆栈信息记录的日志级别
CustomHandler   ：自定义日志处理函数，返回true时，继续执行打印程序，返回false时，不再执行打印程序_
```
- #### FileOption介绍

- ###### FileOption为接口，有FileSizeMode与FileTimeMode两个实现对象
- ###### FileTimeMode对象
    ```text
    Filename   日志文件路径
    Timemode   按小时，天，月份：MODE_HOUR，MODE_DAY，MODE_MONTH
    Maxbuckup  最多备份日志文件数
    IsCompress  备份文件是否压缩
    ```

- ###### FileSizeMode对象
    ```text
	Filename   日志文件路径
	Maxsize    日志文件大小的最大值，超过则滚动备份
	Maxbuckup  最多备份日志文件数
	IsCompress  备份文件是否压缩
    ```

- ##### SetOption示例1
    ```go
    SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileTimeMode{Filename: "testlogtime.log", Maxbuckup: 10, IsCompress: true, Timemode: MODE_DAY}})
    ```
- ##### SetOption示例2
    ```go
    SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileSizeMode{Filename: "testlog.log", Maxsize: 1<<30, Maxbuckup: 10, IsCompress: true}})
    ```
- ##### Stacktrace 堆栈日志
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
  [DEBUG]2024/08/07 18:46:12 logging_test.go:152 this is a debug message
  [INFO]2024/08/07 18:46:12 logging_test.go:157 this is a info message
  [WARN]2024/08/07 18:46:12 logging_test.go:162#logging_test.go:158#logging_test.go:153#testing.go:1689#asm_amd64.s:1695 this is a warn message
  [ERROR]2024/08/07 18:46:12 logging_test.go:167#logging_test.go:163#logging_test.go:158#logging_test.go:153#testing.go:1689#asm_amd64.s:1695 this is a error message
  ```

- #### CustomHandler 自定义函数，可以自定义处理逻辑
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

#### 控制台日志设置
```go
//全局log：
SetConsole(false)  //控制台不打日志,默认值true
//实例log：
log.SetConsole(false)  //控制台不打日志,默认值true
```
***

### 打印日志示例：
```go
//SetRollingFile("", "log.txt", 1000, KB)  设置日志文件信息
//SetRollingFileLoop(``, "log.txt", 300, MB, 50)   设置日志文件大小300M，最多保留50个最近的日志文件
//SetRollingByTime(``, "log.txt", MODE_MONTH) 按月份分割日志
//SetRollingByTime(``, "log.txt", MODE_HOUR)  按小时分割日志
//SetRollingByTime(``, "log.txt", MODE_DAY)  按天分割日志与调用SetRollingDaily("", "log.txt") 作用相同


//控制台不打印
// SetConsole(false)

Debug("00000000000")
//默认格式：[DEBUG]2023/07/10 18:40:49 logging_test.go:12: 00000000000

SetFormat(FORMAT_NANO) 
Debug("111111111111")
//设置格式(无格式化)：111111111111

SetFormat(FORMAT_LONGFILENAME) 
Info("22222222")
//设置格式(长文件路径) ：[INFO]/usr/log/logging/logging_test.go:14: 22222222

SetFormat(FORMAT_DATE | FORMAT_SHORTFILENAME) 
Warn("333333333")
//设置格式(日期+短文件路径) ：[WARN]2023/07/10 logging_test.go:16: 333333333

SetFormat(FORMAT_DATE | FORMAT_TIME) /
Error("444444444")
//设置格式 ：[ERROR]2023/07/10 18:35:19 444444444

SetFormat(FORMAT_SHORTFILENAME)
Fatal("5555555555")
//设置格式 ：[FATAL]logging_test.go:21: 5555555555

SetFormat(FORMAT_TIME)
Fatal("66666666666")
//设置格式 ：[FATAL]18:35:19 66666666666
```

### 校正打印时间
```go
//修改TIME_DEVIATION可以校正日志打印时间，单位纳秒
TIME_DEVIATION 
```

### SetLevelOption  给不同日志级别设置不同的输出日志格式

###### 通过SetLevelOption函数，可以给指定的日志级别设置独立的日志输出格式

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
------

#### 性能测试：

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
5. **Slog 和 goLogger 结合**：指同时使用go-logger作为slog的日志文件管理库。


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

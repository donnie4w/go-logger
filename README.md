## go-logger 高性能golang日志库 [[English]](https://github.com/donnie4w/go-logger/blob/master/README_en.md)

------------

### 性能特点

1. **极高并发性能**：极高的并发写数据性能，比官方库或同类型日志库高**10倍或以上**
2. **极低内存占用**：是官方库与同类型日志库的几分之一

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

-  **日志文件管理**: 支持 直接作为go 标准库  `log/slog`  的日志文件管理器，实现 `slog`的日志文件按小时，天，月份，文件大小等多种方式进行日志文件切割，同时也支持按文件大小切分日志文件后,压缩归档日志文件。
- **一致的性能表现**: `go-logger` + slog 内存分配与性能 与 slog直接写日志文件一致。
- 详细参见[使用文档](https://tlnet.top/logdoc "使用文档")

### [使用文档](https://tlnet.top/logdoc "使用文档")

------------

### 一. 设置日志打印格式 (`SetFormat`)

##### 如： SetFormat(FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME)

##### 默认格式:  `FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME`

```text
不格式化，只打印日志内容		FORMAT_NANO		无格式
长文件名及行数			FORMAT_LONGFILENAME	全路径
短文件名及行数			FORMAT_SHORTFILENAME	如：logging_test.go:10
相对路径文件名及行数              FORMAT_RELATIVEFILENAME 如：logger/test/logging_test.go:10
精确到日期			FORMAT_DATE		如：2023/02/14
精确到秒			        FORMAT_TIME		如：01:33:27
精确到微秒			FORMAT_MICROSECONDS	如：01:33:27.123456
日志级别标识                     FORMAT_LEVELFLAG        如：[Debug],[Info],[Warn][Error][Fatal]             
调用函数                         FORMAT_FUNC             调用函数的函数名，若设置，则出现在文件名之后
```
#### 示例：

```go
logger.SetFormat(logger.FORMAT_LEVELFLAG | logger.FORMAT_LONGFILENAME | logger.FORMAT_TIME)
logger.Error("错误信息：文件未找到")
// 输出:
// [ERROR]/usr/log/logging/main.go:20 10:45:00: 错误信息：文件未找到
```

### 二. 设置日志标识输出格式  (`SetFormatter`)

######  `SetFormatter("{level} {time} {file} {message}\n")`

##### 默认格式：`"{level} {time} {file} {message}\n"`

```text
{level}        日志级别信息：如[Debug],[Info],[Warn],[Error],[Fatal]
{time}         日志时间信息
{file}         文件位置行号信息
{message}      日志内容
```

##### 说明：除了关键标识  `{message}`  `{time}`  `{file}`  `{level}` 外，其他内容原样输出，如 | ， 空格，换行  等

###### 通过修改 `formatter`，可以自由定义输出格式，例如：

```go
logger.SetFormatter("{time} - {level} - {file} - {message}\n")
logger.Info("日志初始化完成")
// 输出:
// 2023/08/09 10:30:00 - [INFO] - main.go:12 - 日志初始化完成
```

------------

### 三. 日志级别 (`SetLevel`)  (`SetLevelOption`)

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

### 四. 文件日志管理

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

### 五. Option参数 （`SetOption`）

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
Formatter       ：日志输出  默认："{level}{time} {file} {mesaage}\n"
FileOption      ：日志文件接口参数
Stacktrace      ：开启日志堆栈信息记录的日志级别
CustomHandler   ：自定义日志处理函数，返回true时，继续执行打印程序，返回false时，不再执行打印程序
AttrFormat      ：日志属性格式化
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

  - ###### `FileMixedMode` 按文件大小和时间混合模式滚动备份日志文件
    ```text
    Filename   日志文件路径
    Maxsize    日志文件大小的最大值，超过则滚动备份
    Timemode   按小时，天，月份：MODE_HOUR，MODE_DAY，MODE_MONTH
    Maxbuckup  最多备份日志文件数
    IsCompress  备份文件是否压缩
    ```

  - ##### SetOption 示例1  `FileTimeMode`
    ```go
    // debug级别，关闭控制台日志打印，按天备份日志，最多日志文件数位10，备份时压缩文件，日志文件名为 testlogtime.log
    SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileTimeMode{Filename: "testlogtime.log", Maxbuckup: 10, IsCompress: true, Timemode: MODE_DAY}})
    ```
  - ##### SetOption 示例2  `FileSizeMode`
    ```go
    // debug级别，关闭控制台日志打印，按文件大小备份日志，按每文件大小为1G时备份一个文件，  最多日志文件数位10，备份时压缩文件，日志文件名为 testlog.log
    SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileSizeMode{Filename: "testlog.log", Maxsize: 1<<30, Maxbuckup: 10, IsCompress: true}})
    ```

  - ##### SetOption 示例3  `FileMixedMode`
    ```go
    // debug级别，关闭控制台日志打印，按天同时按文件大小备份日志，最多日志文件数位10，备份时压缩文件，日志文件名为 mixedlog.log
    SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileSizeMode{Filename: "mixedlog.log", Maxsize: 1<<30, Maxbuckup: 10, IsCompress: true, Timemode: MODE_DAY}})
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
    [FATAL]2024/08/07 20:22:40 logging_test.go:Stacktrace3:167#logging_test.go:Stacktrace2:162#logging_test.go:Stacktrace1:157#logging_test.go:TestStacktrace:152#testing.go:tRunner:1689#asm_amd64.s:goexit:1695 this is a fatal message
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

4. #### `AttrFormat` 日志属性自定义函数

```go
func Test_AttrFormat(t *testing.T) {
	attrformat := &logger.AttrFormat{
		SetLevelFmt: func(level logger.LEVELTYPE) string {
			switch level {
			case logger.LEVEL_DEBUG:
				return "debug:"
			case logger.LEVEL_INFO:
				return "info:"
			case logger.LEVEL_WARN:
				return "warn:"
			case logger.LEVEL_ERROR:
				return "error>>>>"
			case logger.LEVEL_FATAL:
				return "[fatal]"
			default:
				return "[unknown]"
			}
		},
		SetTimeFmt: func() (string, string, string) {
			s := time.Now().Format("2006-01-02 15:04:05")
			return s, "", ""
		},
	}
	logger.SetOption(&logger.Option{AttrFormat: attrformat, Console: true, FileOption: &logger.FileTimeMode{Filename: "testlogtime.log", Maxbuckup: 3, IsCompress: false, Timemode: logger.MODE_MONTH}})
	logger.Debug("this is a debug message:", 1111111111111111111)
	logger.Info("this is a info message:", 2222222222222222222)
	logger.Warn("this is a warn message:", 33333333333333333)
	logger.Error("this is a error message:", 4444444444444444444)
	logger.Fatal("this is a fatal message:", 555555555555555555)
}
```


------------

### 六. 控制台日志设置 (`SetConsole`)
```go
//全局log：
SetConsole(false)  //控制台不打日志,默认值true
//实例log：
log.SetConsole(false)  //控制台不打日志,默认值true
```
***

### 七. 校正打印时间  (`TIME_DEVIATION`)
###### 有时在分布式环境中，可能存在不同机器时间不一致的问题，`go-logger` 允许通过 `TIME_DEVIATION` 参数来进行时间校正。

```go
logger.TIME_DEVIATION = 1000 // 将日志时间校正 +1微妙
```

------

## 性能压测数据： （详细数据可以参考[使用文档](https://tlnet.top/logdoc)）

| 日志记录器                | 核心数 | 操作次数     | 每操作耗时(ns) | 内存分配(B) | 分配次数 |
|-------------------------|-------|------------|--------------|------------|--------|
| Serial_NativeLog        | 4     | 598,425    | 4,095        | 248        | 2      |
| Serial_NativeLog        | 8     | 589,526    | 4,272        | 248        | 2      |
| Serial_Zap              | 4     | 485,172    | 4,943        | 352        | 6      |
| Serial_Zap              | 8     | 491,910    | 4,851        | 353        | 6      |
| Serial_GoLogger         | 4     | 527,454    | 3,987        | 80         | 2      |
| Serial_GoLogger         | 8     | 574,303    | 4,083        | 80         | 2      |
| Serial_Slog             | 4     | 498,553    | 4,952        | 344        | 6      |
| Serial_Slog             | 8     | 466,743    | 4,942        | 344        | 6      |
| Serial_SlogAndGoLogger  | 4     | 443,798    | 5,149        | 344        | 6      |
| Serial_SlogAndGoLogger  | 8     | 460,762    | 5,208        | 344        | 6      |
| Parallel_NativeLog      | 4     | 424,681    | 5,176        | 248        | 2      |
| Parallel_NativeLog      | 8     | 479,988    | 5,045        | 248        | 2      |
| Parallel_Zap            | 4     | 341,937    | 6,736        | 352        | 6      |
| Parallel_Zap            | 8     | 353,247    | 6,517        | 353        | 6      |
| Parallel_GoLogger       | 4     | 4,240,896  | 549.9        | 163        | 3      |
| Parallel_GoLogger       | 8     | 4,441,388  | 550.4        | 128        | 3      |
| Parallel_Slog           | 4     | 477,423    | 4,972        | 344        | 6      |
| Parallel_Slog           | 8     | 447,642    | 5,064        | 344        | 6      |
| Parallel_SlogAndGoLogger| 4     | 424,813    | 5,242        | 345        | 6      |
| Parallel_SlogAndGoLogger| 8     | 425,070    | 5,215        | 345        | 6      |

### 性能分析说明

1. **NativeLog**：go自带log库
2. **Zap**：uber高性能日志库
3. **GoLogger**：go-logger
4. **Slog**：go自带的slog库
5. **SlogAndGoLogger**：使用go-logger作为slog的日志文件管理库


### 性能分析

| 库              | 测试类型       | 并发数 | 平均时间(ns/op) | 内存分配(B/op) | 内存分配次数(allocs/op) |
|------------------|----------------|--------|------------------|------------------|--------------------------|
| **NativeLog**    | Serial         | 4      | 3956             | 248              | 2                        |
|                  |                | 8      | 4044             | 248              | 2                        |
|                  | Parallel       | 4      | 4916             | 248              | 2                        |
|                  |                | 8      | 5026             | 248              | 2                        |
| **Zap**          | Serial         | 4      | 4815             | 352              | 6                        |
|                  |                | 8      | 4933             | 353              | 6                        |
|                  | Parallel       | 4      | 6773             | 352              | 6                        |
|                  |                | 8      | 6610             | 353              | 6                        |
| **GoLogger**     | Serial         | 4      | 4010             | 80               | 2                        |
|                  |                | 8      | 3966             | 80               | 2                        |
|                  | Parallel       | 4      | 568.1            | 165              | 3                        |
|                  |                | 8      | 576.0            | 128              | 3                        |
| **slog**         | Serial         | 4      | 4914             | 344              | 6                        |
|                  |                | 8      | 4921             | 344              | 6                        |
|                  | Parallel       | 4      | 4952             | 344              | 6                        |
|                  |                | 8      | 5075             | 344              | 6                        |
| **slog + GoLogger** | Serial      | 4      | 5058             | 344              | 6                        |
|                  |                | 8      | 5046             | 344              | 6                        |
|                  | Parallel       | 4      | 5150             | 345              | 6                        |
|                  |                | 8      | 5250             | 345              | 6                        |

### 性能分析

1. **NativeLog（log库）**:
    - **串行性能**: 具有相对较低的延迟（3956 ns/op 和 4044 ns/op），且内存占用较少（248 B/op）。
    - **并行性能**: 在并发测试中，NativeLog 的性能也保持稳定，延迟（4916 ns/op 和 5026 ns/op）仍然低于其他库。

2. **Zap（zap库）**:
    - **串行性能**: Zap 的串行性能稍逊色于 log，延迟略高（4815 ns/op 和 4933 ns/op），并且内存占用较高（352 B/op）。
    - **并行性能**: Zap 在并行测试中表现较差，延迟明显高于其他库，达到 6773 ns/op 和 6610 ns/op，显示出其在高并发情况下的瓶颈。

3. **GoLogger（go-logger）**:
    - **串行性能**: 在串行性能上表现良好，延迟（4010 ns/op 和 3966 ns/op），内存使用最低（80 B/op）。
    - **并行性能**: 在并行测试中表现优异，延迟显著低于其他库，仅为 568.1 ns/op 和 576.0 ns/op，显示了其极高的并发处理能力。

4. **slog（slog库）**:
    - **串行性能**: slog 的串行性能在所有库中属于中等水平（4914 ns/op 和 4921 ns/op），内存占用相对较高（344 B/op）。
    - **并行性能**: 在并行情况下，slog 的性能表现中规中矩（4952 ns/op 和 5075 ns/op）。

5. **slog + GoLogger（slog+go-logger）**:
    - **串行性能**: 当结合 slog 和 GoLogger 时，性能表现为（5058 ns/op 和 5046 ns/op），内存占用（344 B/op）与单独使用slog库相同。
    - **并行性能**: 并行测试中，组合使用的性能（5150 ns/op 和 5250 ns/op）。

-------

### 高并发场景中，go-logger的性能比同类型的日志库高10倍以上

##### 在高并发场景中，`go-logger` 的性能显著优于其他日志库，尤其是在处理大量并发日志写入时。以下是根据你提供的基准测试数据，分析的各个日志库的性能：

| 库名                  | 并发性能（ns/op）           | 内存分配（B/op） | 内存分配次数（allocs/op） | 备注           |
|---------------------|-----------------------|-------------------|----------------------------|--------------|
| **NativeLog**       | 4916 - 5026           | 248               | 2                          | Go自带日志库，性能中等 |
| **Zap**             | 6610  - 6773      | 352               | 6                          | 性能一般，适合常规场景  |
| **GoLogger**        | **568.1** - **576.0** | 165               | 3                          | 极高性能，适合高并发场景 |
| **Slog**            | 4952 - 5075           | 344               | 6                          | 性能一般，适合常规使用  |
| **SlogAndGoLogger** | 5150 - 5250           | 345               | 6                          | 与单独使用Slog类似  |

### 分析

1. **GoLogger(go-logger)**:
    - 在高并发环境下，其性能表现极为出色，延迟在 `568.1 ns/op` 和 `576.0 ns/op` 之间，耗时远低于其他库。
    - 内存分配显著更少（165 B/op），这意味着在高负载情况下能更有效地管理内存，减少GC压力。

2. **NativeLog(log)**:
    - 性能适中，延迟在 `4916 ns/op` 到 `5026 ns/op` 之间，内存分配较高（248 B/op），在高并发场景下可能导致性能下降。

3. **Zap(zap)**:
    - 性能较低，延迟在 `6773 ns/op` 到 `6610 ns/op` 之间，内存分配较多（352 B/op），适合普通使用场景。

4. **Slog(slog)**:
    - 性能一般，延迟在 `4952 ns/op` 到 `5075 ns/op` 之间，适合普通使用场景。

5. **SlogAndGoLogger(slog+go-logger)**:
    - 性能稍低于log，与单独使用slog类似，适合使用slog库并且需要管理日志文件的场景。

### 结论

##### 在高并发场景中，`go-logger` 的性能几乎是其他库的10倍以上，这是由于其优化的内存管理和更快的日志写入速度。它是处理大量并发日志写入的最佳选择，尤其是在对性能要求极高的应用中，推荐优先考虑使用 `go-logger`。

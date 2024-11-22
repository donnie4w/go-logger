## go-logger: A high-performance logging library for Golang [[中文]](https://github.com/donnie4w/go-logger/blob/master/README.md)

------------

### Performance Highlights

1. **Exceptional Concurrency**: Handles concurrent writes with performance **over 10 times** faster than the standard library and similar logging libraries.
2. **Minimal Memory Usage**: Utilizes a fraction of the memory required by the standard library and comparable logging libraries.

### Key Features

1. **Log Level Configuration**: Dynamically adjust log levels to control the verbosity of logs across different environments.
2. **Customizable Formatting**: Flexible log formatting options for timestamps, log levels, file location, and more.
3. **File Rotation**: Supports automatic log file rotation based on file count, ensuring manageable log storage.
4. **File Compression**: Archives log files through compression.
5. **Standard Library Log Management**: Supports file splitting, compression, and other management features for the Go `log` and `slog` libraries.
6. **Custom Handlers**: Allows for external handlers to process log output in custom ways.
7. **Stack Trace Logging**: Records the complete call stack, showing file names, function names, and line numbers for each call step.
8. **Level-Specific Log Formatting**: Configure independent output formats for different log levels.

### `go-logger` + `slog`

- **File Management**: Can manage `slog` log files with flexible rotation options based on hours, days, months, or file size, with optional compression.
- **Consistent Performance**: Maintains memory allocation and performance consistent with direct `slog` file writes.
- For detailed usage, see the [documentation](https://tlnet.top/logdoc "Documentation").

### [Documentation](https://tlnet.top/logdoc "Documentation")

------------

### 1. Setting Log Output Format (`SetFormat`)

##### Example: SetFormat(FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME)

##### Default Format: `FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME`

```plaintext
No formatting, logs message content only    FORMAT_NANO             No format
Long filename and line number               FORMAT_LONGFILENAME     Full path
Short filename and line number              FORMAT_SHORTFILENAME    e.g., logging_test.go:10
Relative path and line number               FORMAT_RELATIVEFILENAME e.g., logger/test/logging_test.go:10
Date precision                              FORMAT_DATE             e.g., 2023/02/14
Second precision                            FORMAT_TIME             e.g., 01:33:27
Microsecond precision                       FORMAT_MICROSECONDS     e.g., 01:33:27.123456
Log level indicator                         FORMAT_LEVELFLAG        e.g., [Debug],[Info],[Warn][Error][Fatal]             
Function name                               FORMAT_FUNC             Function name appears after filename if set
```

#### Example:

```go
logger.SetFormat(logger.FORMAT_LEVELFLAG | logger.FORMAT_LONGFILENAME | logger.FORMAT_TIME)
logger.Error("Error: File not found")
// Output:
// [ERROR]/usr/log/logging/main.go:20 10:45:00: Error: File not found
```

### 2. Setting Log Identifier Output Format (`SetFormatter`)

###### `SetFormatter("{level} {time} {file} {message}\n")`

##### Default Format: `"{level} {time} {file} {message}\n"`

```plaintext
{level}      Log level info: e.g., [Debug],[Info],[Warn],[Error],[Fatal]
{time}       Log timestamp
{file}       File and line info
{message}    Log content
```

##### Note: Elements `{message}`, `{time}`, `{file}`, `{level}` are recognized; all other characters (e.g., | , spaces, newlines) are output as-is.

###### Modify `formatter` to define custom formats, for instance:

```go
logger.SetFormatter("{time} - {level} - {file} - {message}\n")
logger.Info("Logger initialized")
// Output:
// 2023/08/09 10:30:00 - [INFO] - main.go:12 - Logger initialized
```

------------

### 3. Log Levels (`SetLevel`, `SetLevelOption`)

##### Log Level Order: DEBUG < INFO < WARN < ERROR < FATAL

###### Disable all logs with `SetLevel(OFF)`

`go-logger` supports a wide range of log levels, from `DEBUG` to `FATAL`, configurable with `SetLevel` to control output verbosity:

```go
logger.SetLevel(logger.INFO)
logger.Debug("Debug info: this will not be logged")
logger.Info("Info: this will be logged")
```

##### Additionally, use `SetLevelOption` to set distinct formats for each log level:

```go
logger.SetLevelOption(logger.LEVEL_DEBUG, &logger.LevelOption{
    Format: logger.FORMAT_SHORTFILENAME | logger.FORMAT_TIME,
})
logger.SetLevelOption(logger.LEVEL_WARN, &logger.LevelOption{
    Format: logger.FORMAT_LONGFILENAME | logger.FORMAT_DATE | logger.FORMAT_FUNC,
})
```

##### Example: Setting distinct formats for different log levels

```go
func TestLevelOptions(t *testing.T) {
  SetLevelOption(LEVEL_DEBUG, &LevelOption{Format: FORMAT_LEVELFLAG | FORMAT_TIME | FORMAT_SHORTFILENAME})
  SetLevelOption(LEVEL_INFO, &LevelOption{Format: FORMAT_LEVELFLAG})
  SetLevelOption(LEVEL_WARN, &LevelOption{Format: FORMAT_LEVELFLAG | FORMAT_TIME | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_FUNC})
  
  Debug("This is a debug message")
  Info("This is an info message")
  Warn("This is a warning message")
}
```
##### Output:
```plaintext
[DEBUG]18:53:55 logging_test.go:176 This is a debug message
[INFO]This is an info message
[WARN]2024/08/07 18:53:55 logging_test.go:TestLevelOptions:178 This is a warning message
```

### 4. Log File Management

##### `go-logger` supports file logging with various rotation strategies and backup options.

###### To use the global log object, call configuration methods directly:
```go
SetRollingDaily()     Rotate by day
SetRollingByTime()    Rotate by hour, day, or month
SetRollingFile()      Rotate based on file size
SetRollingFileLoop()  Rotate by size with a set maximum number of files
SetGzipOn(true)       Enable log file compression
```
#### Multiple Instances:
```go
log1 := NewLogger()
log1.SetRollingDaily("", "logMonitor.log")

log2 := NewLogger()
log2.SetRollingDaily("", "logBusiness.log")
```
#### 1. Rotate Log Files by Date
```go
log.SetRollingDaily("/var/logs", "log.txt")
// Daily log files named like log_20221015.txt, and if it exists, will append .1, .2, etc.

log.SetRollingByTime("/var/logs", "log.txt",MODE_MONTH)
// Monthly rotation keeps previous months' logs, e.g.:
// log_202210.txt, log_202211.txt, log_202212.txt

log.SetRollingByTime("/var/logs", "log.txt",MODE_HOUR)
// Rotate hourly, e.g.:
// log_2022101506.txt, log_2022101507.txt, log_2022101508.txt
```
#### 2. Rotate Log Files by Size
```go
log.SetRollingFile("/var/logs", "log.txt", 300, MB)
// Backup when file exceeds 300MB as log.1.txt, log.2.txt, etc.

log.SetRollingFileLoop("/var/logs", "log.txt", 300, MB, 50)
// Set max size to 300MB with up to 50 recent log files kept
```

------

### 5. Option Parameters (`SetOption`)

###### `Option` parameters provide flexible configurations, organizing log settings for clearer maintenance.

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

##### Properties:
```plaintext
Level           : Log level
Console         : Print to console
Format          : Log format, default: FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
Formatter       : Log output format, default: "{level}{time} {file} {message}\n"
FileOption      : Log file settings
Stacktrace      : Stack trace logging level
CustomHandler   : Custom log handler function; return true to continue, false to skip log entry
AttrFormat      : Custom attribute formatting
```

1. FileOption Overview

- ###### FileOption is an interface with two implementation classes: `FileSizeMode` and `FileTimeMode`.
- ###### `FileTimeMode` rotates log files based on time.
  ```text
  Filename   Log file path
  Timemode   Rotation interval by hour, day, or month: MODE_HOUR, MODE_DAY, MODE_MONTH
  Maxbackup  Maximum number of log file backups
  IsCompress Whether the backup file is compressed
  ```

- ###### `FileSizeMode` rotates log files based on file size.
  ```text
  Filename   Log file path
  Maxsize    Maximum log file size; rotation occurs when size is exceeded
  Maxbackup  Maximum number of log file backups
  IsCompress Whether the backup file is compressed
  ```

- ###### `FileMixedMode` rotates log files based on file size and time.
  ```text
  Filename   Log file path
  Timemode   Rotation interval by hour, day, or month: MODE_HOUR, MODE_DAY, MODE_MONTH
  Maxsize    Maximum log file size; rotation occurs when size is exceeded
  Maxbackup  Maximum number of log file backups
  IsCompress Whether the backup file is compressed
  ```



- ##### SetOption Example 1
  ```go
  // debug level, disable console log printing, daily log rotation, maximum of 10 log files, compress backups, log file named testlogtime.log
  SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileTimeMode{Filename: "testlogtime.log", Maxbackup: 10, IsCompress: true, Timemode: MODE_DAY}})
  ```

- ##### SetOption Example 2
  ```go
  // debug level, disable console log printing, rotate log files by file size, rotate at 1G per file, maximum of 10 log files, compress backups, log file named testlog.log
  SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileSizeMode{Filename: "testlog.log", Maxsize: 1<<30, Maxbackup: 10, IsCompress: true}})
  ```

- ##### SetOption Example 3
  ```go
  // debug level, disable console log printing, rotate log files by file size and time,  maximum of 10 log files, compress backups, log file named mixedlog.log
  SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileSizeMode{Filename: "mixedlog.log", Maxsize: 1<<30, Maxbackup: 10, IsCompress: true, Timemode: MODE_DAY}})
  ```

2. Stacktrace Log

- ###### The stacktrace log feature records the entire function call sequence from the log point to the program entry, including each step's file name, function name, and line number. This is highly useful for debugging and error analysis.
- When the log level is `WARN` or higher, the log entry will include full call stack information.
- **Example**

  ```go
  func TestStacktrace(t *testing.T) {
      SetOption(&Option{Console: true, Stacktrace: LEVEL_WARN, Format: FORMAT_LEVELFLAG | FORMAT_DATE | FORMAT_TIME | FORMAT_SHORTFILENAME | FORMAT_FUNC})
      Debug("this is a debug message")
      Stacktrace1()
  }

  func Stacktrace1() {
      Info("this is an info message")
      Stacktrace2()
  }

  func Stacktrace2() {
      Warn("this is a warn message")
      Stacktrace3()
  }

  func Stacktrace3() {
      Error("this is an error message")
      Fatal("this is a fatal message")
  }
  ```

  ##### Execution Result
  ```go
  [DEBUG]2024/08/07 20:22:40 logging_test.go:TestStacktrace:151 this is a debug message
  [INFO]2024/08/07 20:22:40 logging_test.go:Stacktrace1:156 this is an info message
  [WARN]2024/08/07 20:22:40 logging_test.go:Stacktrace2:161#logging_test.go:Stacktrace1:157#logging_test.go:TestStacktrace:152#testing.go:tRunner:1689#asm_amd64.s:goexit:1695 this is a warn message
  [ERROR]2024/08/07 20:22:40 logging_test.go:Stacktrace3:166#logging_test.go:Stacktrace2:162#logging_test.go:Stacktrace1:157#logging_test.go:TestStacktrace:152#testing.go:tRunner:1689#asm_amd64.s:goexit:1695 this is an error message
  [FATAL]2024/08/07 20:22:40 logging_test.go:Stacktrace3:167#logging_test.go:Stacktrace2:162#logging_test.go:Stacktrace1:157#logging_test.go:TestStacktrace:152#testing.go:tRunner:1689#asm_amd64.s:goexit:1695 this is a fatal message
  ```

3. `CustomHandler` - Custom Function for Handling Logic

- **Example**
  ```go
  func TestCustomHandler(t *testing.T) {
      SetOption(&Option{Console: true, CustomHandler: func(lc *LogContext) bool {
          fmt.Println("level:", levelname(lc.Level))
          fmt.Println("message:", fmt.Sprint(lc.Args...))
          if lc.Level == LEVEL_ERROR {
              return false // Do not print if it's an error message
          }
          return true
      },
      })
      Debug("this is a debug message")
      Info("this is an info message")
      Warn("this is a warn message")
      Error("this is an error message")
  }
  ```
- ##### Execution Result: Error logs are not printed based on `CustomHandler` logic.
  ```go
  level: debug
  message: this is a debug message
  [DEBUG]2024/08/07 18:51:56 logging_test.go:126 this is a debug message
  level: info
  message: this is an info message
  [INFO]2024/08/07 18:51:56 logging_test.go:127 this is an info message
  level: warn
  message: this is a warn message
  [WARN]2024/08/07 18:51:56 logging_test.go:128 this is a warn message
  level: error
  message: this is an error message
  ```

4. `AttrFormat` - Custom Attribute Format Function for Logs

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
        logger.SetOption(&logger.Option{AttrFormat: attrformat, Console: true, FileOption: &logger.FileTimeMode{Filename: "testlogtime.log", Maxbackup: 3, IsCompress: false, Timemode: logger.MODE_MONTH}})
        logger.Debug("this is a debug message:", 1111111111111111111)
        logger.Info("this is an info message:", 2222222222222222222)
        logger.Warn("this is a warn message:", 33333333333333333)
        logger.Error("this is an error message:", 4444444444444444444)
        logger.Fatal("this is a fatal message:", 555555555555555555)
    }
    ```

---

### 6. Console Log Setting (`SetConsole`)
```go
// Global log:
SetConsole(false)  // Disable console logging, default value is true
// Instance log:
log.SetConsole(false)  // Disable console logging, default value is true
```

### 7. Adjusting Log Print Time (`TIME_DEVIATION`)
###### In distributed environments, time inconsistencies may occur across machines. `go-logger` allows time adjustment using the `TIME_DEVIATION` parameter.

```go
logger.TIME_DEVIATION = 1000 // Adjust log time by +1 microsecond
```
-------

## Performance Benchmark Data: (Detailed data can be referenced in the [Usage Documentation](https://tlnet.top/logdoc))

| Logger                  | Core Count | Operations  | Time per Op (ns) | Memory Allocation (B) | Allocations |
|-------------------------|------------|-------------|-------------------|-----------------------|-------------|
| Serial_NativeLog        | 4          | 598,425     | 4,095            | 248                   | 2           |
| Serial_NativeLog        | 8          | 589,526     | 4,272            | 248                   | 2           |
| Serial_Zap              | 4          | 485,172     | 4,943            | 352                   | 6           |
| Serial_Zap              | 8          | 491,910     | 4,851            | 353                   | 6           |
| Serial_GoLogger         | 4          | 527,454     | 3,987            | 80                    | 2           |
| Serial_GoLogger         | 8          | 574,303     | 4,083            | 80                    | 2           |
| Serial_Slog             | 4          | 498,553     | 4,952            | 344                   | 6           |
| Serial_Slog             | 8          | 466,743     | 4,942            | 344                   | 6           |
| Serial_SlogAndGoLogger  | 4          | 443,798     | 5,149            | 344                   | 6           |
| Serial_SlogAndGoLogger  | 8          | 460,762     | 5,208            | 344                   | 6           |
| Parallel_NativeLog      | 4          | 424,681     | 5,176            | 248                   | 2           |
| Parallel_NativeLog      | 8          | 479,988     | 5,045            | 248                   | 2           |
| Parallel_Zap            | 4          | 341,937     | 6,736            | 352                   | 6           |
| Parallel_Zap            | 8          | 353,247     | 6,517            | 353                   | 6           |
| Parallel_GoLogger       | 4          | 4,240,896   | 549.9            | 163                   | 3           |
| Parallel_GoLogger       | 8          | 4,441,388   | 550.4            | 128                   | 3           |
| Parallel_Slog           | 4          | 477,423     | 4,972            | 344                   | 6           |
| Parallel_Slog           | 8          | 447,642     | 5,064            | 344                   | 6           |
| Parallel_SlogAndGoLogger| 4          | 424,813     | 5,242            | 345                   | 6           |
| Parallel_SlogAndGoLogger| 8          | 425,070     | 5,215            | 345                   | 6           |

### Performance Analysis specification

1. **NativeLog**: Go's built-in logging library
2. **Zap**: Uber’s high-performance logging library
3. **GoLogger**: go-logger
4. **Slog**: Go’s built-in Slog library
5. **Slog + GoLogger**: Slog using GoLogger for log file management

### Performance Analysis

| Library           | Test Type        | Concurrency | Avg. Time (ns/op) | Mem Allocation (B/op) | Allocations (allocs/op) |
|-------------------|------------------|-------------|--------------------|------------------------|--------------------------|
| **NativeLog**     | Serial           | 4           | 3956              | 248                    | 2                        |
|                   |                  | 8           | 4044              | 248                    | 2                        |
|                   | Parallel         | 4           | 4916              | 248                    | 2                        |
|                   |                  | 8           | 5026              | 248                    | 2                        |
| **Zap**           | Serial           | 4           | 4815              | 352                    | 6                        |
|                   |                  | 8           | 4933              | 353                    | 6                        |
|                   | Parallel         | 4           | 6773              | 352                    | 6                        |
|                   |                  | 8           | 6610              | 353                    | 6                        |
| **GoLogger**      | Serial           | 4           | 4010              | 80                     | 2                        |
|                   |                  | 8           | 3966              | 80                     | 2                        |
|                   | Parallel         | 4           | 568.1             | 165                    | 3                        |
|                   |                  | 8           | 576.0             | 128                    | 3                        |
| **Slog**          | Serial           | 4           | 4914              | 344                    | 6                        |
|                   |                  | 8           | 4921              | 344                    | 6                        |
|                   | Parallel         | 4           | 4952              | 344                    | 6                        |
|                   |                  | 8           | 5075              | 344                    | 6                        |
| **Slog + GoLogger** | Serial         | 4           | 5058              | 344                    | 6                        |
|                   |                  | 8           | 5046              | 344                    | 6                        |
|                   | Parallel         | 4           | 5150              | 345                    | 6                        |
|                   |                  | 8           | 5250              | 345                    | 6                        |

### Performance Analysis

1. **NativeLog (log)**:
    - **Serial Performance**: Offers relatively low latency (3956 ns/op and 4044 ns/op) with low memory usage (248 B/op).
    - **Parallel Performance**: Performance remains stable under parallel testing with latency (4916 ns/op and 5026 ns/op), still lower than other libraries.

2. **Zap (zap)**:
    - **Serial Performance**: Slightly lower than log, with higher latency (4815 ns/op and 4933 ns/op) and higher memory usage (352 B/op).
    - **Parallel Performance**: Performs worse in parallel testing, with latency peaking at 6773 ns/op and 6610 ns/op, highlighting limitations under high concurrency.

3. **GoLogger (go-logger)**:
    - **Serial Performance**: Performs well with latency (4010 ns/op and 3966 ns/op), and the lowest memory usage (80 B/op).
    - **Parallel Performance**: Excellent parallel performance with the lowest latency at 568.1 ns/op and 576.0 ns/op, showing high concurrency capabilities.

4. **Slog (slog)**:
    - **Serial Performance**: Average performance among all libraries, with latency (4914 ns/op and 4921 ns/op) and higher memory usage (344 B/op).
    - **Parallel Performance**: Consistent performance with latency (4952 ns/op and 5075 ns/op) under parallel testing.

5. **Slog + GoLogger (slog+go-logger)**:
    - **Serial Performance**: Combined performance (5058 ns/op and 5046 ns/op) with memory usage (344 B/op) similar to slog alone.
    - **Parallel Performance**: Slightly lower performance than log alone, making it suitable when using slog with additional log management from GoLogger.

-------

### GoLogger demonstrates 10x higher performance than similar libraries in high-concurrency environments

##### GoLogger shows a significant advantage in handling high-concurrency log writing, with benchmark data revealing its superiority over other logging libraries:

| Library         | Concurrency Performance (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Notes                  |
|-----------------|---------------------------------|---------------------------|------------------------------|------------------------|
| **NativeLog**   | 4916 - 5026                     | 248                       | 2                            | Go’s default logger, medium performance |
| **Zap**         | 6610 - 6773                     | 352                       | 6                            | Moderate performance, suited for general use |
| **GoLogger**    | **568.1 - 576.0**               | 165                       | 3                            | Superior performance, ideal for high-concurrency |
| **Slog**        | 4952 - 5075                     | 344                       | 6                            | Average performance, suitable for regular use |
| **Slog+GoLogger** | 5150 - 5250                   | 345                       | 6                            | Similar to standalone Slog |

### Analysis

1. **GoLogger (go-logger)**:
    - Outstanding performance in high-concurrency scenarios, with latency of `568.1 ns/op` and `576.0 ns/op`, much faster than other libraries.
    - Lower memory allocations (165 B/op) reduce GC pressure, optimizing memory management under high load.

2. **NativeLog (log)**:
    - Moderate performance with latency between `4916 ns/op` and `5026 ns/op`, higher memory usage (248 B/op) may

3. **Zap (zap)**:
    - Has the most consistent performance in single-threaded environments but suffers significant latency in high-concurrency scenarios (latency peaks at 6773 ns/op).

4. **Slog (slog)**:
    - General performance with latency between `4952 ns/op` and `5075 ns/op`, suitable for standard usage scenarios, providing stable results without high concurrency optimizations.

5. **SlogAndGoLogger (slog+go-logger)**:
    - Slightly lower performance than log, similar to using slog alone; suitable when combining slog’s standardized logging with GoLogger’s file management functionality, ideal for use cases prioritizing both features and performance.

### Conclusion

##### In high-concurrency scenarios, `go-logger` performs nearly 10 times faster than other libraries, thanks to its optimized memory management and faster logging speed. It is the optimal choice for handling large volumes of concurrent log writes, particularly in applications with strict performance requirements, and is highly recommended for such cases.
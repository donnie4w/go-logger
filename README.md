## go-logger 是go 高性能日志库

------------

### 功能

- 日志级别设置：允许动态调整日志级别，以便在不同环境下控制日志的详细程度。
- 格式化输出：支持自定义日志的输出格式，包括时间戳、日志级别、日志位置 等元素。
- 文件数回滚：支持按照日志文件数自动文件回滚，并防止日志文件数过多。
- 文件压缩：支持压缩归档日志文件。
- 支持标准库log/slog日志文件管理：支持标准库文件切割，压缩等功能。

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
	精确到秒				FORMAT_TIME		如：01:33:27
	精确到微秒			FORMAT_MICROSECNDS	如：01:33:27.123456

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

	使用全局log对象时，直接调用设置方法：

	SetRollingDaily()		按日期分割
	SetRollingByTime()		可按 小时，天，月 分割日志
	SetRollingFile()		指定文件大小分割日志
	SetRollingFileLoop()		指定文件大小分割日志，并指定保留最大日志文件数
	SetGzipOn(true)			压缩分割的日志文件 

#### 多实例：

	log1 := NewLogger()
	log1.SetRollingDaily("", "logMonitor.log")
	 
	log12:= NewLogger()
	log2.SetRollingDaily("", "logBusiness.log")

#### 1. 按日期分割日志文件

	log.SetRollingDaily("d:/foldTest", "log.txt")
	每天按 log_20221015.txt格式 分割
	若 log_20221015.txt已经存在，则生成 log_20221015.1.txt ，log_20221015.2.txt等文件
	
	log.SetRollingByTime("d:/foldTest", "log.txt",MODE_MONTH)
	按月份分割日志，跨月时，保留上月份日志，如：
		log_202210.txt
		log_202211.txt
		log_202212.txt
	
	log.SetRollingByTime("d:/foldTest", "log.txt",MODE_HOUR)
	按小时分割日志, 如：
		log_2022101506.txt
		log_2022101507.txt
		log_2022101508.txt

#### 2. 按文件大小分割日志文件

	log.SetRollingFile("d:/foldTest", "log.txt", 300, MB)
	当文件超过300MB时，按log.1.txt，log.2.txt 格式备份
	目录参数可以为空，则默认当前目录。
	
	log.SetRollingFileLoop(`d:/foldTest`, "log.txt", 300, MB, 50) 
	设置日志文件大小最大为300M
	日志文件只保留最新的50个

### Option参数

###### 建议通过option参数设置log的所有参数，更加易于维护

	Level      ：日志级别
	Console    ：控制台打印
	Format     ：日志格式，默认：FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
	Formatter  ：日志输出  默认："{level} {time} {file}:{mesaage}\n"
	FileOption ：日志文件接口参数

#### FileOption介绍

###### FileOption为接口，有FileSizeMode与FileTimeMode两个实现对象

###### FileTimeMode对象

	Filename   日志文件路径
	Timemode   按小时，天，月份：MODE_HOUR，MODE_DAY，MODE_MONTH
	Maxbuckup  最多备份日志文件数
	IsCompress  备份文件是否压缩

###### FileSizeMode对象

	Filename   日志文件路径
	Maxsize    日志文件大小的最大值，超过则滚动备份
	Maxbuckup  最多备份日志文件数
	IsCompress  备份文件是否压缩

##### SetOption示例1

	SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileTimeMode{Filename: "testlogtime.log", Maxbuckup: 10, IsCompress: true, Timemode: MODE_DAY}})

##### SetOption示例2

	SetOption(&Option{Level: LEVEL_DEBUG, Console: false, FileOption: &FileSizeMode{Filename: "testlog.log", Maxsize: 1<<30, Maxbuckup: 10, IsCompress: true}})


------------

#### 控制台日志设置

	全局log：SetConsole(false)控制台不打日志,默认值true
	实例log：log.SetConsole(false)控制台不打日志,默认值true

***

### 打印日志示例：

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
	/设置格式 ：[ERROR]2023/07/10 18:35:19 444444444
	
	SetFormat(FORMAT_SHORTFILENAME)
	Fatal("5555555555")
	//设置格式 ：[FATAL]logging_test.go:21: 5555555555

	SetFormat(FORMAT_TIME)
	Fatal("66666666666")
	//设置格式 ：[FATAL]18:35:19 66666666666

#### 校正打印时间

	//修改TIME_DEVIATION可以校正日志打印时间，单位纳秒
	TIME_DEVIATION 


#### 性能测试：

###### 测试说明

|  测试日志库 |  描述|
| ------------ | ------------ |
|  zap |	"go.uber.org/zap" 高性能日志库常规格式化输出	   |
|  go-logger | go-logger 常规格式化输出  |
| go-logger NoFORMAT  |  go-logger 无格式化输出 |
|  go-logger write |  go-logger write方法写数据 |
|slog   |  go 原生 slog库 |

##### 测试数据1

###### 测试环境

**amd64 cpu: Intel(R) Core(TM) i5-1035G1 CPU @ 1.00GHz**

![](https://tlnet.top/f/1696141149_1696133036.jpg)

![](https://tlnet.top/f/1696141691_1696133161.jpg)

![](https://tlnet.top/f/1696141697_1696133275.jpg)

![](https://tlnet.top/f/1696141701_1696133381.jpg)

##### 测试结果

###### 时间消耗
- go-logger    4500ns/op 左右
- slog与zap    5600ns/op 左右

###### 内存消耗
- go-logger  64
- slog与zap  330  左右

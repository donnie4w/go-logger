### go-logger 是golang 的日志库 ，基于对golang内置log的封装。
**在控制台打印：直接调用 Debug()，Info()，Warn(), Error() ,Fatal() 日志级别由低到高**
级别概念 类似java日志工具log4j

## **设置日志打印格式：**
如： SetFormat(FORMAT_SHORTFILENAME|FORMAT_DATE|FORMAT_TIME)<br>

	无其他格式，只打印日志内容	FORMAT_NANO		无格式
	长文件名及行数			FORMAT_LONGFILENAME		全路径
	短文件名及行数			FORMAT_SHORTFILENAME	如：logging_test.go:10
	精确到日期			FORMAT_DATE		如：2023/02/14
	精确到秒				FORMAT_TIME		如：01:33:27
	精确到微秒			FORMAT_MICROSECNDS
打印结果形如：[DEBUG]2023/02/14 01:33:27 logging_test.go:10: 11111111111111 <br>
若需要**自定义格式** 只需要 SetFormat(FORMAT_NANO) ，既可以去掉原有格式。<br>

## **日志级别**
**ALL < DEBUG < INFO < WARN < ERROR < FATAL < OFF** <br>
默认日志级别为ALL，说明： <br>

	若设置 INFO
	如：SetLevel(INFO)
	则 所有 Debug("*********")   不再打印出来
	所以调试阶段，常设置为默认级别ALL，或DEBUG，打印出项目中所有日志，包括调试日志
	若设置 OFF
	SetLevel(OFF)
	则 所有日志不再打印出来
	所以正式环境，常设置为ERROR或以上的日志级别，项目中Debug()，Info(),warn()等日志不再打印出来，具体视实际需求设置
	


**需将日志写入文件时，则要设置日志文件名**<br>
    使用全局对象log时，直接调用设置方法：

	SetRollingDaily()或SetRollingFile()
需要**多实例**指定不同日志文件时：<br>

	log1 := logger.NewLogger()
	log1.SetRollingDaily("", "logMonitor.log")
	 
	log12:= logger.NewLogger()
	log2.SetRollingDaily("", "logBusiness.log")
    

**1. 按日期分割日志文件**

	log.SetRollingDaily("d://foldTest", "log.txt")
	每天按 log_20221015.txt格式 分割
	若 log_20221015.txt已经存在，则生成 log_20221015.1.txt ，log_20221015.2.txt等文件

**2. 按文件大小分割日志文件**

	log.SetRollingFile("d://foldTest", "log.txt", 300, MB)
	按文件超过300MB是，按log.1.txt，log.2.txt 格式备份
	目录参数可以为空，则默认当前目录。

**控制台日志设置**

	全局log：SetConsole(false)控制台不打日志,默认值true
	实例log：log.SetConsole(false)控制台不打日志,默认值true

***

### 打印日志示例：

	// SetRollingFile("", "log.txt", 1000, KB)  设置日志文件信息
	// SetConsole(false)  控制台打印信息，默认true
	Debug("11111111")
	Info("22222222")
	SetFormat(FORMAT_DATE | FORMAT_SHORTFILENAME) //设置后，下面日志格式只打印日期+短文件信息
	Warn("333333333")
	SetLevel(FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	Error("444444444")
	Fatal("5555555555")
	
	/*获取新的log实例，要求不同日志文件时，使用多实例对象*/
	log := NewLogger()
	/*按日期分割日志文件*/
	//log.SetRollingDaily("", "log.txt")	
	/*按日志文件大小分割日志文件*/
	// log.SetRollingFile("", "log.txt", 3, MB)
	/* 设置打印级别 OFF,DEBUG,INFO,WARN,ERROR,FATAL*/
	//log.SetLevel(ALL) 默认ALL.

	/* 日志写入文件时，同时在控制台打印出来，设置为false后将不打印在控制台，默认值true*/
	// log.SetConsole(false)
	log.Debug("aaaaaaaaaaaaa")
	log.SetFormat(FORMAT_LONGFILENAME) //设置后将打印出文件全部路径信息
	log.Info("bbbbbbbbbbbb")
	log.SetFormat(FORMAT_MICROSECNDS | FORMAT_SHORTFILENAME)//设置日志格式，时间+短文件名
	log.Warn("ccccccccccccccc")
	log.SetLevel(FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	log.Error("dddddddddddd")
	log.Fatal("eeeeeeeeeeeee")

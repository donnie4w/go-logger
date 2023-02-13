### go-logger 是golang 的日志库 ，基于对golang内置log的封装。
**在控制台打印：直接调用 Debug()，Info()，Warn(), Error() ,Fatal() 日志级别由低到高**
级别概念 类似java日志工具log4j

**设置日志打印格式：**
如： SetFormat(FORMAT_SHORTFILENAME|FORMAT_DATE|FORMAT_TIME)
	无其他格式，只打印日志内容
	FORMAT_NANO
	长文件名及行数
	FORMAT_LONGFILENAME
	短文件名及行数
	FORMAT_SHORTFILENAME
	精确到日期
	FORMAT_DATE
	精确到秒
	FORMAT_TIME
	精确到微秒
	FORMAT_MICROSECNDS

**需写日志文件时，可以获取实例**
    全局单实例可以直接调用        log := logging.GetStaticLogger() 
    要求多实例可获取新实例可以调用 log := logging.NewLogger()
**1. 按日期分割日志文件**
    	log.SetRollingDaily("d://foldTest", "log.txt")
	每天按 log_20221015.txt格式备份
**2. 按文件大小分割日志文件**
	log.SetRollingFile("d://foldTest", "log.txt", 300, MB)
	按文件超过300MB是，按log.1.txt，log.2.txt 格式备份
**控制台**
	log.SetConsole(false)控制台不打日志,默认值true
  
***

### 打印日志示例：
**控制台打印，直接调用打印方法Debug(),Info()等方法**
	Debug("11111111111111")
	Info("22222222")
	SetFormat(FORMAT_DATE | FORMAT_SHORTFILENAME)//设置后，下面日志格式只打印日期+短文件信息
	Warn("333333333")
	SetLevel(FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	Error("444444444")
	Fatal("5555555555")


**设置日志文件示例**

	/*获取全局log单例，单日志文件项目日志建议使用单例*/
	//log := GetStaticLogger()

	/*获取新的log实例，要求不同日志文件时，使用多实例对象*/
	log := NewLogger()

	/*按日期分割日志文件，也是默认设置值*/
	log.SetRollingDaily("d://cfoldTest", "log.txt")
	/*按日志文件大小分割日志文件*/
	// log.SetRollingFile("d://cfoldTest", "log.txt", 3, MB)

	/* 设置打印级别 OFF,DEBUG,INFO,WARN,ERROR,FATAL
	log.SetLevel(OFF) 设置OFF后，将不再打印后面的日志 默认日志级别为ALL，打印级别*/

	/* 日志写入文件时，同时在控制台打印出来，设置为false后将不打印在控制台，默认值true*/
	// log.SetConsole(false)

	log.Debug("aaaaaaaaaaaaa")
	log.SetFormat(FORMAT_LONGFILENAME) //设置后将打印出文件全部路径信息
	log.Info("bbbbbbbbbbbb")
	log.SetFormat(FORMAT_MICROSECNDS | FORMAT_SHORTFILENAME)//设置日志格式，时间+短文件名
	log.Warn("cccccccccccc")
	log.SetLevel(FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	log.Error("dddddddddddd")
	log.Fatal("eeeeeeeeeeee")

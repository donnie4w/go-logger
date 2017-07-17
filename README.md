### go-logger 是golang 的日志库 ，基于对golang内置log的封装。
用法类似java日志工具包log4j

**打印日志有5个方法 Debug，Info，Warn, Error ,Fatal  日志级别由低到高**

设置日志级别的方法为：logger.SetLevel() 如：logger.SetLevel(logger.WARN)
则：logger.Debug(....),logger.Info(...) 日志不会打出，而 
 logger.Warn(...),logger.Error(...),logger.Fatal(...)日志会打出。
设置日志级别的参数有7个，分别为：ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF
其中 ALL表示所有调用打印日志的方法都会打出，而OFF则表示都不会打出。

***

日志文件切割有两种类型：1为按日期切分。2为按日志大小切分。
按日期切分时：每天一个备份日志文件，后缀为 .yyyy-MM-dd 
过0点是生成前一天备份文件

按大小切分是需要3个参数，1为文件大小，2为单位，3为文件数量
文件增长到指定限值时，生成备份文件，结尾为依次递增的自然数。
文件数量增长到指定限制时，新生成的日志文件将覆盖前面生成的同名的备份日志文件。

**示例**：

	//指定是否控制台打印，默认为true
	logger.SetConsole(true)
	//指定日志文件备份方式为文件大小的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	//第三个参数为备份文件最大数量
	//第四个参数为备份文件大小
	//第五个参数为文件大小的单位 KB，MB，GB TB
	//logger.SetRollingFile("d:/logtest", "test.log", 10, 5, logger.KB)

	//指定日志文件备份方式为日期的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	logger.SetRollingDaily("d:/logtest", "test.log")

	//指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高
	//一般习惯是测试阶段为debug，生成环境为info以上
	logger.SetLevel(logger.DEBUG)


### 打印日志：
func log(i int) {
	logger.Debug("Debug>>>>>>>>>>>>>>>>>>>>>>" , strconv.Itoa(i))
	logger.Info("Info>>>>>>>>>>>>>>>>>>>>>>>>>" , strconv.Itoa(i))
	logger.Warn("Warn>>>>>>>>>>>>>>>>>>>>>>>>>" , strconv.Itoa(i))
	logger.Error("Error>>>>>>>>>>>>>>>>>>>>>>>>>, strconv.Itoa(i))
	logger.Fatal("Fatal>>>>>>>>>>>>>>>>>>>>>>>>>",  strconv.Itoa(i))
}


## v1.0.3
### 增加
**1.输出日志格式化，与 log.Printf(format,"***")  format类型，如 **
  1. logger.SetFormat("%s,%s")  对应参数是两个：logger.Debug("a","b")  输出：a,b 

**2.支持多日志对象，不同对象可以设置不同存放日志目录或日志名称**
  1. 通用配置 ：logger.SetRollingFile() ，logger.SetRollingDaily()
  2. 指定对象1   		   log1 := logger.GetLogger()   log1.SetRollingDaily("d:/logtest", "t1.log")
  3. 指定对象2   		   log2 := logger.GetLogger()   log2.SetRollingDaily("d:/logtest", "t2.log")

**3.支持同一对象指定日志级别对应不同文件**
  1. logger.SetLevelFile(level LEVEL, fileDir, fileName string)
  2. 参数说明：level 日志级别 ；fileDir 日志文件夹；fileName  日志文件名
  3. 如：logger.SetLevelFile(logger.INFO, "d:/logtest", "info.log")  
  4. 如：logger.SetLevelFile(logger.WARN, "d:/logtest", "warn.log")

### 修改
  1. 重构代码
  2. 优化，去掉监控器



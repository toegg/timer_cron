package main

import(
	"fmt"
	"time"
	"strconv"
	"strings"
)

//时间类型
const (
	NOTHING   = 0		//不匹配不执行
	ANYTHING  = 1		//"*"
	NUM       = 2		//具体数字
	RANGE     = 3		//范围
)

//范围
type Range struct{
	first int
	last  int
	step  int
}

//字段规则
type Field struct{
	type_data int 	//时间类型
	num  int  		//时间类型为NUM, 的具体数字
	range_val Range //时间类型为RANGE, 的范围
}

//规则
type Entry struct{
	M 	Field
	H   Field
	D   Field
	MN  Field
	W   Field
	DoFunc func()
}

//当前时间
type NowTime struct{
	Minute int
	Hour   int
	Day    int
	Month  int
	Week   int		
}

//规则列表
var EntryList []Entry

var exit = make(chan int)

//过期时间,零点加上2s
var time_out = 60 + 2

//初始化
func init(){
	parse_file()	
}

//解析规则列表
func parse_file() {
	//对应分，时，日，月，周几
	add_entrys("10", "*", "*", "*", "*", func(){fmt.Println("every 10 minute handler")}) //第10分执行

	add_entrys("0-10/2", "*", "*", "*", "*", func(){fmt.Println("every 0-10 minute 2 step handler")}) //第0-10分中，每2分钟执行

	add_entrys("50-55", "*", "*", "*", "*", func(){fmt.Println("every 50 - 55 minute handler")}) //第50-55分执行

    add_entrys("*/1", "*", "*", "*", "*", func(){fmt.Println("one minute handler")}) //每分钟执行

    add_entrys("0", "*/1", "*", "*", "*", func(){fmt.Println("one hour handler")}) //每小时执行
}

func main(){
	second := time.Now().Second()
	first := time.NewTimer(time.Duration(62 - second) * time.Second)
	//首次要预约到下个零点执行后再进入轮询
	select {
		case <-first.C:
			timer_do(EntryList)
			cron := time.NewTicker(60 * time.Second)
			go start_cron(cron)
	}
	<- exit
}

//开启定时轮询
func start_cron(cron *time.Ticker){
	for {
		select {
			case <-cron.C:
				fmt.Println("timer_do")
				timer_do(EntryList)
		}
	}	
}

//增加规则
//m 分
//h 时
//d 日
//mn 月
//w 周几
//func function函数
func add_entrys(m string, h string, d string, mn string, w string, fun func()) {
	entry := Entry{
		M:parse_entry(1, m),
		H:parse_entry(2, h),
		D:parse_entry(3, d),
		MN:parse_entry(4, mn),
		W:parse_entry(5, w),
		DoFunc:fun,
	}
	EntryList = append(EntryList, entry)
}

func parse_entry(pos int, val string) Field{
	if pos == 1 {
		return parse_field(val, 0, 59)
	}
	if pos == 2 {
		return parse_field(val, 0, 23)
	}
	if pos == 3 {
		return parse_field(val, 1, 31)
	}
	if pos == 4 {
		return parse_field(val, 1, 12)
	}
	if pos == 5 {
		return parse_field(val, 1, 7)
	}	

	return Field{type_data:NOTHING}
}

func parse_field(val string, val1 int, val2 int ) Field{
	if val == "*" {
		return Field{type_data:ANYTHING}
	} 
	if strings.Contains(val, "/") == false && strings.Contains(val, "-") == false {
		num, _ := strconv.Atoi(val)
		return Field{type_data:NUM, num:num}
	}else{
		var range_str string
		var step int

		args1 := strings.Split(val, "/")
		if len(args1) == 1 {
			range_str = args1[0]
			step = 1
		}else{
			range_str = args1[0]
			step, _ = strconv.Atoi(args1[1])
		}
		args2 := strings.Split(range_str, "-")
		if len(args2) == 1 {
			range_val := Range{first:val1, last:val2, step:step}
			return Field{type_data:RANGE, range_val:range_val}
		}else{
			first, _ := strconv.Atoi(args2[0])
			last, _ := strconv.Atoi(args2[1])
			range_val := Range{first:first, last:last, step:step}
			return Field{type_data:RANGE, range_val:range_val}
		}
	}

	return Field{type_data:NOTHING}
}

//轮询
func timer_do(entrys []Entry) {
	now_time := get_time()
    for _, entry := range entrys{
    	if can_run(entry, now_time) {
    		entry.DoFunc()
    	}
    }
}

//检测规则是否符合
func can_run(entry Entry, now_time NowTime) bool {
	return field_ok(entry.M, now_time.Minute) &&
		field_ok(entry.H, now_time.Hour) &&
		field_ok(entry.D, now_time.Day) &&
		field_ok(entry.MN, now_time.Month) &&
		field_ok(entry.W, now_time.Week)
}

//字段检测
func field_ok(match_val Field, val int) bool{
	if match_val.type_data == NOTHING {
		return false
	}

	if match_val.type_data == ANYTHING {
		return true
	}

	if match_val.type_data == NUM {
		return match_val.num == val
	}

	if match_val.type_data == RANGE {
		return range_ok(match_val.range_val, val)
	}

	return false
}

//检测范围是否符合
func range_ok(match_val Range, val int) bool{
	if match_val.first > val || match_val.last < val {
		return false
	}

	Result := range_ok1(match_val.first, match_val.last, match_val.step, val)
	return Result
}

func range_ok1(val1 int, val2 int, val3 int, val int) bool {
	if val1 > val2 {
		return false
	}
	if val1 == val {
		return true
	}
	return range_ok1(val1 + val3, val2, val3, val)
}

//获取当前时间
func get_time() NowTime{
	time := time.Now()
    minute := time.Minute()
    hour := time.Hour()
    day := time.Day()
    month := int(time.Month())
    week := int(time.Weekday())
	return NowTime{Minute:minute, Hour:hour, Day:day, Month:month, Week:week}
}
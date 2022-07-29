# noGcStaticMap

对于大型map，比如总数达到千万级别的map,如果键或者值中包含引用类型(string类型，结构体类型，或者任何基本类型+指针的定义 *int, *float 等)，那么这个map在垃圾回收的时候就会非常慢，GC的周期回收时间可以达到秒级甚至分钟级。

对此参考fastcache等，把复杂的不利于GC的复杂map转化为基础类型的map map[uint64]int 用于存储索引 和 []byte用于存储实际键值。如此改造之后，基本上实现了零GC。此noGcStaticMap无hash碰撞问题，可以放心使用。 

使用限制：只适用于单次加载完后，键值对不再变化的情况。即在键值加载完成之前，只允许新增，不允许查询;在键值对加载完成后，则只能查询，不能再新增或删除键值。
```go
package main

import (
	"github.com/yudeguang/noGcStaticMap"
	"log"
	"strconv"
)
//声明成全局变量
var m1 = noGcStaticMap.NewDefault()
var m2 = noGcStaticMap.NewInt()

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	tAny()
	tInt()
}

func tAny() {
	log.Println("开始")

	//增加
	m1.Set([]byte(""), []byte("键为空的值"))               //键为空
	m1.Set([]byte(strconv.Itoa(1000000)), []byte("")) //值为空
	for i := 0; i < 1000; i++ {
		m1.Set([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}
	//加载完成 加载完成后不允许再加载 未加载完成前，不允许查询
	m1.SetFinished()
	//查询键为空
	val, exist := m1.GetString("")
	log.Println("key:", "", "值:", val, exist)
	//查询键为空
	val, exist = m1.GetString(strconv.Itoa(1000000))
	log.Println("key:", 1000000, "值:", val, exist)
	for i := 0; i < 10; i++ {
		val, exist = m1.GetString(strconv.Itoa(i))
		log.Println("key:", i, "值:", val, exist)
	}
	log.Println("完成查询")
}
func tInt() {
	log.Println("开始")

	m2.Set(1000000, []byte("")) //值为空
	for i := 0; i < 1000; i++ {
		m2.Set(i, []byte(strconv.Itoa(i)))
	}
	//加载完成 加载完成后不允许再加载 未加载完成前，不允许查询
	m2.SetFinished()
	//查询空值
	val, exist := m2.GetString(1000000)
	log.Println("key:", 1000000, "值:", val, exist)
	//查询普通值
	for i := 0; i < 10; i++ {
		val, exist := m2.GetString(i)
		log.Println("key:", i, "值:", val, exist)
	}
	log.Println("完成查询")
}

```

# noGcStaticMap

对于大型map，比如总数达到千万级别的map,如果键或者值中包含引用类型(string类型，结构体类型，或者任何基本类型+指针的定义 *int, *float 等)，那么这个MAP在垃圾回收的时候就会非常慢，GC的周期回收时间可以达到秒级。

所以对于这种map需要进行优化，把复杂的不利于GC的复杂map转化为基础类型的  map[uint64]int+ []byte，变成这种形式之后，整个gc基本就不耗时了。

注意，这种做法主要适用于单次加载完后，键值对不再变化的情况。对于键值对在运行过程中还要动态增减的情况则不适合。
```go
package main

import (
	"github.com/yudeguang/noGcStaticMap"
	"log"
	"strconv"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	tAny()
	tInt()
}

func tAny() {
	log.Println("开始")
	n := noGcStaticMap.New()
	//增加
	n.Set([]byte(""), []byte("空"))                    //键为空
	n.Set([]byte(strconv.Itoa(1000000)), []byte("空")) //值为空
	for i := 0; i < 1000; i++ {
		n.Set([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}
	//加载完成 加载完成后不允许再加载 未加载完成前，不允许查询
	n.SetFinished()
	//查询键为空
	val, exist := n.GetString([]byte(""))
	log.Println("key:", "", "值:", val, exist)
	//查询键为空
	val, exist = n.GetString([]byte(strconv.Itoa(1000000)))
	log.Println("key:", 1000000, "值:", val, exist)
	for i := 0; i < 10; i++ {
		val, exist = n.GetString([]byte(strconv.Itoa(i)))
		log.Println("key:", i, "值:", val, exist)
	}
	log.Println("完成查询")
}
func tInt() {
	log.Println("开始")
	n := noGcStaticMap.NewInt()
	n.Set(1000000, []byte("")) //值为空
	for i := 0; i < 1000; i++ {
		n.Set(i, []byte(strconv.Itoa(i)))
	}
	//加载完成 加载完成后不允许再加载 未加载完成前，不允许查询
	n.SetFinished()
	//查询空值
	val, exist := n.GetString(1000000)
	log.Println("key:", 1000000, "值:", val, exist)
	//查询普通值
	for i := 0; i < 10; i++ {
		val, exist := n.GetString(i)
		log.Println("key:", i, "值:", val, exist)
	}
	log.Println("完成查询")
}
```

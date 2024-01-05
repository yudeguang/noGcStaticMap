# noGcStaticMap

https://github.com/yudeguang/noGcMap 与 https://github.com/yudeguang/noGcStaticMap 为同一系列的无GC类型MAP，两者针对的场景有一定差异,noGcStaticMap性能稍高，内存占用更小，但不支持增删改。

对于大型map，比如总数达到千万级别的map,如果键或者值中包含引用类型(string类型，结构体类型，或者任何基本类型+指针的定义 *int, *float 等)，那么这个map在垃圾回收的时候就会非常慢，GC的周期回收时间可以达到秒级甚至分钟级。

对此参考fastcache等，把复杂的不利于GC的复杂map转化为基础类型的map map[uint64]uint32 用于存储索引 和 []byte用于存储实际键值。如此改造之后，基本上实现了零GC,总体而言：

优点:

1)几乎零GC;

2)无hash碰撞问题;

3)内存占用相对较小;

4)提供GetUnsafe,GetValFromDataBeginPosOfKVPairUnSafe等函数以满足高性能场景的要求(不复制内容，直接取值);

5)代码量非常少，适合根据自己需求做二次修改;


缺点:

1)为纯静态map，不能动态新增或删除键值对,即在键值加载完成之前，只允许新增;在键值对加载完成后，则只允许查询;

注意：

对于一些结构体类型，把结构体与[]byte的相互转换，可能会用到convert_help.go 文件中的StructToStr,SliceToStr以及BytesToStruct,BytesToSlice等函数，这些函数需要自己复制后改写实现。 


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

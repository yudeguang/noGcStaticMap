// Copyright 2022 rateLimit Author(https://github.com/yudeguang/noGcStaticMap). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/yudeguang/noGcStaticMap.

package noGcStaticMap

import (
	"fmt"
	"github.com/yudeguang/iox"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
)

// 结构体案例
type NoGcStructExample struct {
	Col1 int
	Col2 string
	Col3 string
	Col4 string
}

// 缓存池1024
var bytePoolForConvert1024 = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 1024)
		return b
	},
}

// 缓存池10240
var bytePoolForConvert10240 = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 10240)
		return b
	},
}

// 缓存池102400
var bytePoolForConvert102400 = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 102400)
		return b
	},
}

// 缓存池1024000
var bytePoolForConvert1024000 = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 1024000)
		return b
	},
}

// 缓存池10240000
var bytePoolForConvert10240000 = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 10240000)
		return b
	},
}

// 默认的文本分隔符，一般文本中不会有这个字符
var SplitSep = []byte("`")

// 高效的对数据进行分割，相比于用系统的Split减少gc的产生,速度更快
func BytesToStruct(data []byte) (p NoGcStructExample) {
	dataLen := len(data)
	dataBeginPos := 0
	if dataLen == 0 {
		panic("data err")
	}
	//因为要经常分配内存，所以要用缓存池,并且要用合理的缓存池大小
	var byteBuf []byte
	curFieldLocation := 1
	if dataLen < 1024-4 {
		byteBuf = bytePoolForConvert1024.Get().([]byte)
	} else if dataLen >= 1024-4 && dataLen < 10240-4 {
		byteBuf = bytePoolForConvert10240.Get().([]byte)
	} else if dataLen >= 10240-4 && dataLen < 102400-4 {
		byteBuf = bytePoolForConvert102400.Get().([]byte)
	} else if dataLen >= 102400-4 && dataLen < 1024000-4 {
		byteBuf = bytePoolForConvert1024000.Get().([]byte)
	} else if dataLen >= 1024000-4 && dataLen < 10240000-4 {
		byteBuf = bytePoolForConvert10240000.Get().([]byte)
	} else {
		panic("data is too long")
	}
	var byteBufDataLen int
	for i := 0; i < dataLen; i++ {
		//注意i == int(valLen)-1 的问题
		if data[dataBeginPos+i] == SplitSep[0] || i == dataLen-1 {
			if i == dataLen-1 {
				byteBuf[byteBufDataLen] = data[dataBeginPos+i]
				byteBufDataLen = byteBufDataLen + 1
			}
			curField := string(byteBuf[0:byteBufDataLen])
			//读取到SplitSep就换一个值，并且根据结构体的实际情况转换为对应的类型
			switch curFieldLocation {
			case 1:
				col, err := strconv.Atoi(curField)
				if err != nil {
					panic(err)
				}
				p.Col1 = col
			case 2:
				p.Col2 = curField
			case 3:
				p.Col3 = curField
			case 4:
				p.Col4 = curField
			default:
				panic(curField)
			}
			//清空此byteBuf,通过移动offset实现
			byteBufDataLen = 0
			curFieldLocation = curFieldLocation + 1
		} else {
			//数据添加到byteBuf，并移动offset
			byteBuf[byteBufDataLen] = data[dataBeginPos+i]
			byteBufDataLen = byteBufDataLen + 1
		}
	}
	//byteBuf用完，返回给系统
	if dataLen < 1024-4 {
		bytePoolForConvert1024.Put(byteBuf)
	} else if dataLen >= 1024-4 && dataLen < 10240-4 {
		bytePoolForConvert10240.Put(byteBuf)
	} else if dataLen >= 10240-4 && dataLen < 102400-4 {
		bytePoolForConvert102400.Put(byteBuf)
	} else if dataLen >= 102400-4 && dataLen < 1024000-4 {
		bytePoolForConvert1024000.Put(byteBuf)
	} else if dataLen >= 1024000-4 && dataLen < 10240000-4 {
		bytePoolForConvert10240000.Put(byteBuf)
	} else {
		panic("data is too long")
	}
	return p
}

// 数据转换为切片
func BytesToSlice(data []byte) (ps []NoGcStructExample) {
	rd := iox.NewReadSeekerFromBytes(data)
	for {
		b, err := rd.ReadBytesUint32()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal("加载失败", err)
			}
		}
		p := BytesToStruct(b)
		ps = append(ps, p)
	}
	return ps
}

// 把结构体切片转换成字符串
func SliceToStr(p []NoGcStructExample) string {
	var strs string
	for i := range p {
		s := StructToStr(p[i])
		strs = strs + string(uint32ToByte(uint32(len(s))))
		strs = strs + s
	}
	return strs
}

// 把结构体转换成字符串
func StructToStr(p NoGcStructExample) string {
	return JoinInterface(string(SplitSep),
		p.Col1,
		checkStr(p.Col2),
		checkStr(p.Col3),
		checkStr(p.Col4))
}
func JoinInterface(sep string, elems ...interface{}) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(elems[0])
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(fmt.Sprint(elems[i]))
	}
	var b strings.Builder
	b.Grow(n)
	b.WriteString(fmt.Sprint(elems[0]))
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(fmt.Sprint(s))
	}
	return b.String()
}

// 防止文本字段中有分隔符
func checkStr(s string) string {
	if strings.Contains(s, string(SplitSep)) {
		panic(s + "can't contains " + string(SplitSep) + " Please modify the SplitSep")
	}
	return s
}

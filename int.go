// Copyright 2022 rateLimit Author(https://github.com/yudeguang/noGcStaticMap). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/yudeguang/noGcStaticMap.
package noGcStaticMap

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type NoGcStaticMapInt struct {
	setFinished  bool //是否完成存储
	dataBeginPos int  //游标，记录位置
	len          int  //记录键值对个数
	bw           *bufio.Writer
	tempFile     *os.File            //硬盘上的临时文件
	tempFileName string              //临时文件名
	data         []byte              //存储值的内容
	index        [512]map[int]uint32 //值为切片data []byte中的某个位置
}

//初始化 键的类型为int,值的最大长度为65535，与默认类型相比，速度稍快，稍微节省存储空间
func NewInt() *NoGcStaticMapInt {
	var n NoGcStaticMapInt
	for i := range n.index {
		n.index[i] = make(map[int]uint32)
	}
	//创建用于读写的临时文件 同一程序中同时初始化，可能会产生时间相同的问题，需要保证文件名唯一
	var err error
	for {
		n.tempFileName = strconv.Itoa(int(time.Now().UnixNano())) + ".NoGcStaticMap"
		if fileExist(n.tempFileName) {
			continue
		} else {
			break
		}
	}
	n.tempFile, err = os.Create(n.tempFileName)
	haserrPanic(err)
	n.bw = bufio.NewWriterSize(n.tempFile, 40960)
	return &n
}

//取出数据
func (n *NoGcStaticMapInt) Get(k int) (v []byte, exist bool) {
	if !n.setFinished {
		panic("cant't Get before SetFinished")
	}
	idx := k % 512
	dataBeginPos, exist := n.index[idx][k]

	if exist {
		return n.read(int(dataBeginPos)), true
	}
	return v, false
}

//取出数据 以引用的方式取出 对于取出的值，要小心使用
func (n *NoGcStaticMapInt) GetUnsafe(k int) (v []byte, exist bool) {
	dataBeginPos, exist := n.GetDataBeginPosOfKVPair(k)
	if !exist {
		return nil, false
	}
	return n.GetValFromDataBeginPosOfKVPairUnSafe(int(dataBeginPos)), true
}

//取出数据,以string的方式
func (n *NoGcStaticMapInt) GetString(k int) (v string, exist bool) {
	vbyte, exist := n.Get(k)
	if exist {
		return string(vbyte), true
	}
	return v, false
}

//取出键值对在数据中存储的开始位置
func (n *NoGcStaticMapInt) GetDataBeginPosOfKVPair(k int) (uint32, bool) {
	if !n.setFinished {
		panic("cant't Get before SetFinished")
	}
	idx := k % 512
	//这里无需校检键是否正确，故直接返回
	dataBeginPos, exist := n.index[idx][k]
	return dataBeginPos, exist
}

//从内存中的某个位置取出键值对中值的数据
//警告:
//1)传入的dataBeginPos必须是真实有效的，否则有可能会数据越界;
//2)返回的数据是hash表中值的引用，而非值的复制品，要注意不要在外部改变该返回值
func (n *NoGcStaticMapInt) GetValFromDataBeginPosOfKVPairUnSafe(dataBeginPos int) (v []byte) {
	//读取键值的长度 写得能懂直接从fastcache复制过来
	kvLenBuf := n.data[dataBeginPos : dataBeginPos+2]
	valLen := (uint64(kvLenBuf[0]) << 8) | uint64(kvLenBuf[1])
	//读取键的内容，并判断键是否相同
	dataBeginPos = dataBeginPos + 2
	//读取值并返回
	if valLen == 0 {
		return nil
	}
	return n.data[dataBeginPos : dataBeginPos+int(valLen)]
}

//增加数据
func (n *NoGcStaticMapInt) Set(k int, v []byte) {
	n.len = n.len + 1
	//键值设置完之后，不允许再添加
	if n.setFinished {
		panic("can't Set after SetFinished")
	}
	idx := k % 512
	//判断键值的长度，不允许太长
	if len(v) > 65535 {
		panic("k or v is too long,The maximum is 65535")
	}

	_, exist := n.index[idx][k]
	if exist {
		panic("can't add the key '" + strconv.Itoa(k) + "' for twice")
	} else {
		n.index[idx][k] = uint32(n.dataBeginPos)
	}
	//存储数据到临时文件，并且移动游标
	n.write(v)
}

//增加数据,以string的方式
func (n *NoGcStaticMapInt) SetString(k int, v string) {
	n.Set(k, []byte(v))
}

//从内存中读取相应数据
func (n *NoGcStaticMapInt) read(dataBeginPos int) (v []byte) {
	//读取键值的长度 写得能懂直接从fastcache复制过来
	kvLenBuf := n.data[dataBeginPos : dataBeginPos+2]
	valLen := (uint64(kvLenBuf[0]) << 8) | uint64(kvLenBuf[1])
	//读取键的内容，并判断键是否相同
	dataBeginPos = dataBeginPos + 2
	//读取值并返回
	if valLen == 0 {
		return nil
	}
	v = make([]byte, 0, int(valLen))
	v = append(v, n.data[dataBeginPos:dataBeginPos+int(valLen)]...)
	return v
}

//往文件中写入数据
func (n *NoGcStaticMapInt) write(v []byte) {
	dataLen := 2 + len(v) //2个字节表示V的长度
	//直接从fastcache复制过来
	var kvLenBuf [2]byte
	kvLenBuf[0] = byte(uint16(len(v)) >> 8)
	kvLenBuf[1] = byte(len(v))
	//写入v的长度
	_, err := n.bw.Write(kvLenBuf[:])
	haserrPanic(err)
	//写入V
	for i := range v {
		err = n.bw.WriteByte(v[i])
		haserrPanic(err)
	}
	//写完了，移动游标
	n.dataBeginPos = n.dataBeginPos + dataLen
}

//完成存储把存储到硬盘上的文件复制到内存
func (n *NoGcStaticMapInt) SetFinished() {
	n.setFinished = true
	err := n.bw.Flush()
	haserrPanic(err)
	if n.tempFile != nil {
		err = n.tempFile.Close()
		haserrPanic(err)
	}
	b, err := ioutil.ReadFile(n.tempFileName)
	haserrPanic(err)
	n.data = make([]byte, 0, len(b))
	n.data = append(n.data, b...)
	err = os.Remove(n.tempFileName)
	haserrPanic(err)
}

//返回键值对个数
func (n *NoGcStaticMapInt) Len() int {
	return n.len
}

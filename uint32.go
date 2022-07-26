// Copyright 2022 rateLimit Author(https://github.com/yudeguang/noGcStaticMap). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/yudeguang/noGcStaticMap.
package noGcStaticMap

import (
	"bufio"
	"github.com/yudeguang/haserr"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type NoGcStaticMapUint32 struct {
	setFinished  bool //是否完成存储
	dataBeginPos int  //游标，记录位置
	bw           *bufio.Writer
	tempFile     *os.File            //硬盘上的临时文件
	tempFileName string              //临时文件名
	data         []byte              //用于实际存储键值的数据
	index        [512]map[uint32]uint32 //值为切片data []byte中的某个位置
}

//初始化
func NewUint32() *NoGcStaticMapUint32 {
	var n NoGcStaticMapUint32
	for i := range n.index {
		n.index[i] = make(map[uint32]uint32)
	}
	//创建用于读写的临时文件
	var err error
	n.tempFileName = strconv.Itoa(int(time.Now().UnixNano())) + ".NoGcStaticMap"
	n.tempFile, err = os.Create(n.tempFileName)
	haserr.Panic(err)
	n.bw = bufio.NewWriterSize(n.tempFile, 40960)
	return &n
}

//取出数据 在
func (n *NoGcStaticMapUint32) Get(k uint32) (v []byte, exist bool) {
	if !n.setFinished {
		panic("cant't Get before SetFinished")
	}
	idx := k % 512
	dataBeginPos, exist := n.index[idx][k]

	if exist {
		return n.read(k, int(dataBeginPos))
	}
	return v, false
}

//取出数据 在
func (n *NoGcStaticMapUint32) GetString(k uint32) (v string, exist bool) {
	if !n.setFinished {
		panic("cant't Get before SetFinished")
	}
	idx := k % 512
	dataBeginPos, exist := n.index[idx][k]
	if exist {
		val, exist := n.read(k, int(dataBeginPos))
		return string(val), exist
	}
	return v, false
}

/*
K,v的长度用uint16表示(65535)
因为一般说来K不可能太长
*/
func (n *NoGcStaticMapUint32) Set(k uint32, v []byte) {
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
		panic("can't add the key '" + strconv.Itoa(int(k)) + "' for twice")
	} else {
		n.index[idx][k] = uint32(n.dataBeginPos)
	}
	//存储数据到临时文件，并且移动游标
	n.write(k, v)
}

//完成存储把存储到硬盘上的文件复制到内存
func (n *NoGcStaticMapUint32) SetFinished() {
	n.setFinished = true
	err := n.bw.Flush()
	haserr.Panic(err)
	if n.tempFile != nil{
		err=n.tempFile.Close()
		haserr.Panic(err)
	}
	b, err := ioutil.ReadFile(n.tempFileName)
	haserr.Panic(err)
	n.data = make([]byte, 0, len(b))
	n.data = append(n.data, b...)
	err = os.Remove(n.tempFileName)
	haserr.Panic(err)
}

//从内存中读取相应数据
func (n *NoGcStaticMapUint32) read(k uint32, dataBeginPos int) (v []byte, exist bool) {
	//读取键值的长度 写得能懂直接从fastcache复制过来
	kvLenBuf := n.data[dataBeginPos : dataBeginPos+2]
	valLen := (uint64(kvLenBuf[0]) << 8) | uint64(kvLenBuf[1])
	//读取键的内容，并判断键是否相同
	dataBeginPos = dataBeginPos + 2
	//读取值并返回
	if valLen == 0 {
		return nil, true
	}
	return n.data[dataBeginPos : dataBeginPos+int(valLen)], true
}

//往文件中写入数据
func (n *NoGcStaticMapUint32) write(k uint32, v []byte) {
	dataLen := 2 + len(v) //2个字节表示V的长度
	//直接从fastcache复制过来
	var kvLenBuf [2]byte
	kvLenBuf[0] = byte(uint16(len(v)) >> 8)
	kvLenBuf[1] = byte(len(v))
	//写入v的长度
	_, err := n.bw.Write(kvLenBuf[:])
	haserr.Panic(err)
	//写入V
	for i := range v {
		err = n.bw.WriteByte(v[i])
		haserr.Panic(err)
	}
	//写完了，移动游标
	n.dataBeginPos = n.dataBeginPos + dataLen
}

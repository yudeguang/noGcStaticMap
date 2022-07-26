// Copyright 2022 rateLimit Author(https://github.com/yudeguang/noGcStaticMap). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/yudeguang/noGcStaticMap.
package noGcStaticMap

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/cespare/xxhash"
	"io/ioutil"
	"os"
)

type NoGcStaticMapAny struct {
	setFinished         bool //是否完成存储
	dataBeginPos        int  //游标，记录位置
	len                 int  //记录键值对个数
	bw                  *bufio.Writer
	tempFile            *os.File               //硬盘上的临时文件
	tempFileName        string                 //临时文件名
	data                []byte                 //存储键值的内容
	index               [512]map[uint64]uint32 //值为切片data []byte中的某个位置,此索引存储无hash冲突的key的hash值以及有hash冲突但是是第1次出现的key的hash值
	mapForHashCollision map[string]uint32      //值为切片data []byte中的某个位置,string为存放有hash冲突的第2次或2次以上出现的key,这个map一般来说是非常小的
}

//初始化 默认类型,键值的最大长度为65535
func NewDefault(tempFileName ...string) *NoGcStaticMapAny {
	var n NoGcStaticMapAny
	n.mapForHashCollision = make(map[string]uint32)
	for i := range n.index {
		n.index[i] = make(map[uint64]uint32)
	}
	n.tempFileName, n.tempFile, n.bw = createTempFile(tempFileName...)
	return &n
}

//取出数据
func (n *NoGcStaticMapAny) Get(k []byte) (v []byte, exist bool) {
	if !n.setFinished {
		panic("cant't Get before SetFinished")
	}
	h := xxhash.Sum64(k)
	idx := h % 512
	dataBeginPos, exist := n.index[idx][h]
	if exist {
		v, exist = n.read(k, int(dataBeginPos))
		if exist {
			return
		}
	}
	//上面没找到，再从可能存在hash冲突的小表查找
	dataBeginPos, exist = n.mapForHashCollision[string(k)]
	if exist {
		return n.read(k, int(dataBeginPos))
	}
	return v, false
}

//取出数据 警告:返回的数据是hash表中值的引用，而非值的复制品，要注意不要在外部改变该返回值
func (n *NoGcStaticMapAny) GetUnsafe(k []byte) (v []byte, exist bool) {
	dataBeginPos, exist := n.GetDataBeginPosOfKVPair(k)
	if !exist {
		return nil, false
	}
	return n.GetValFromDataBeginPosOfKVPairUnSafe(int(dataBeginPos)), true
}

//取出数据,以string的方式
func (n *NoGcStaticMapAny) GetString(k string) (v string, exist bool) {
	vbyte, exist := n.Get([]byte(k))
	if exist {
		return string(vbyte), true
	}
	return v, false
}

//取出键值对在数据中存储的开始位置
func (n *NoGcStaticMapAny) GetDataBeginPosOfKVPair(k []byte) (uint32, bool) {
	if !n.setFinished {
		panic("cant't Get before SetFinished")
	}
	h := xxhash.Sum64(k)
	idx := h % 512
	dataBeginPos, exist := n.index[idx][h]
	if exist {
		_, exist = n.read(k, int(dataBeginPos))
		if exist {
			return dataBeginPos, true
		}
	}
	//上面没找到，再从可能存在hash冲突的小表查找
	dataBeginPos, exist = n.mapForHashCollision[string(k)]
	if exist {
		_, exist = n.read(k, int(dataBeginPos))
		if exist {
			return dataBeginPos, true
		}
	}
	return 0, false
}

//从内存中的某个位置取出键值对中值的数据
//警告:
//1)传入的dataBeginPos必须是真实有效的，否则有可能会数据越界;
//2)返回的数据是hash表中值的引用，而非值的复制品，要注意不要在外部改变该返回值
func (n *NoGcStaticMapAny) GetValFromDataBeginPosOfKVPairUnSafe(dataBeginPos int) (v []byte) {
	//读取键值的长度
	kvLenBuf := n.data[dataBeginPos : dataBeginPos+4]
	keyLen := (uint64(kvLenBuf[0]) << 8) | uint64(kvLenBuf[1])
	valLen := (uint64(kvLenBuf[2]) << 8) | uint64(kvLenBuf[3])
	//读取键的内容，并判断键是否相同
	dataBeginPos = dataBeginPos + 4 + int(keyLen)
	return n.data[dataBeginPos : dataBeginPos+int(valLen)]
}

//增加数据
func (n *NoGcStaticMapAny) Set(k, v []byte) {
	n.len = n.len + 1
	//键值设置完之后，不允许再添加
	if n.setFinished {
		panic("can't Set after SetFinished")
	}
	h := xxhash.Sum64(k)
	idx := h % 512
	//判断键值的长度，不允许太长
	if len(k) > 65535 || len(v) > 65535 {
		panic("k or v is too long,The maximum is 65535")
	}
	//处理hash碰撞问题
	_, exist := n.index[idx][h]
	if exist {
		//尽可能的避免重复加载,如果在mapNoHashCollision加载过，确实也是无法检测的，但是如果加载了3次一定会被检测到
		if _, exist := n.mapForHashCollision[string(k)]; exist {
			panic("can't add the key '" + string(k) + "' for twice")
		}
		n.mapForHashCollision[string(k)] = uint32(n.dataBeginPos)
	} else {
		n.index[idx][h] = uint32(n.dataBeginPos)
	}
	//存储数据到临时文件，并且移动游标
	n.write(k, v)
}

//增加数据,以string的方式
func (n *NoGcStaticMapAny) SetString(k, v string) {
	n.Set([]byte(k), []byte(v))
}

//从内存中读取相应数据
func (n *NoGcStaticMapAny) read(k []byte, dataBeginPos int) (v []byte, exist bool) {
	//读取键值的长度
	kvLenBuf := n.data[dataBeginPos : dataBeginPos+4]
	keyLen := (uint64(kvLenBuf[0]) << 8) | uint64(kvLenBuf[1])
	valLen := (uint64(kvLenBuf[2]) << 8) | uint64(kvLenBuf[3])
	//读取键的内容，并判断键是否相同
	dataBeginPos = dataBeginPos + 4
	if !bytes.Equal(k, n.data[dataBeginPos:dataBeginPos+int(keyLen)]) {
		return v, false
	}
	//读取值并返回
	if valLen == 0 {
		return nil, true
	}
	dataBeginPos = dataBeginPos + int(keyLen)
	v = make([]byte, 0, int(valLen))
	v = append(v, n.data[dataBeginPos:dataBeginPos+int(valLen)]...)
	return v, true
}

//往文件中写入数据
func (n *NoGcStaticMapAny) write(k, v []byte) {
	dataLen := 4 + len(k) + len(v) //前2个字节表示K占用的空间,之后2个字节表示V的长度
	//直接从fastcache复制过来
	var kvLenBuf [4]byte
	kvLenBuf[0] = byte(uint16(len(k)) >> 8)
	kvLenBuf[1] = byte(len(k))
	kvLenBuf[2] = byte(uint16(len(v)) >> 8)
	kvLenBuf[3] = byte(len(v))
	//写入Kv的长度
	_, err := n.bw.Write(kvLenBuf[:])
	haserrPanic(err)
	//写入k
	for i := range k {
		err = n.bw.WriteByte(k[i])
		haserrPanic(err)
	}
	//写入V
	for i := range v {
		err = n.bw.WriteByte(v[i])
		haserrPanic(err)
	}
	//写完了，移动游标
	n.dataBeginPos = n.dataBeginPos + dataLen
}

//完成存储把存储到硬盘上的文件复制到内存
func (n *NoGcStaticMapAny) SetFinished() {
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
func (n *NoGcStaticMapAny) Len() int {
	return n.len
}

//把INT转换成BYTE
func uint32ToByte(num uint32) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.LittleEndian, num)
	haserrPanic(err)
	return buffer.Bytes()
}

// Copyright 2022 rateLimit Author(https://github.com/yudeguang/noGcStaticMap). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/yudeguang/noGcStaticMap.
package noGcStaticMap

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//创建用于读写的临时文件 同一程序中同时初始化，可能会产生时间相同的问题，需要保证文件名唯一
func createTempFile(args ...string) (tempFileName string, tempFile *os.File, bw *bufio.Writer) {
	var err error
	if len(args) > 0 {
		tempFileName = args[0] + ".NoGcStaticMap"
		if fileExist(tempFileName){
			err := os.Remove(tempFileName)
			haserrPanic(err)
		}
		tempFile, err = os.Create(tempFileName)
		haserrPanic(err)
		return tempFileName, tempFile, bufio.NewWriterSize(tempFile, 40960)
	}
	for {
		tempFileName = strconv.Itoa(int(time.Now().UnixNano())) + ".NoGcStaticMap"
		if fileExist(tempFileName) {
			continue
		} else {
			tempFile, err = os.Create(tempFileName)
			haserrPanic(err)
			return tempFileName, tempFile,bufio.NewWriterSize(tempFile, 40960)
		}
	}
}

//检察文件是或者目录否存在
func fileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

//判断有无错误,并返回true或者false,有错误时调用panic退出
func haserrPanic(err error) bool {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		file = file[strings.LastIndex(file, `/`)+1:]
		panic(fmt.Sprintf("%v,第%v行,错误类型:%v", file, line, err))
		return true
	} else {
		return false
	}
}

// Copyright 2022 rateLimit Author(https://github.com/yudeguang/noGcStaticMap). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/yudeguang/noGcStaticMap.
package noGcStaticMap

import (
	"strconv"
	"testing"
)

func TestHuge(t *testing.T) {
	var mapOffical = make(map[string]string)
	var m = NewHuge("mapHugeForTest")
	//set
	for i := 0; i < 10000; i++ {
		mapOffical[strconv.Itoa(i)] = strconv.Itoa(i)
		m.SetString(strconv.Itoa(i), strconv.Itoa(i))
	}
	m.SetFinished()
	//log.Println(m.GetDataBeginPosOfKVPair([]byte(strconv.Itoa(0))))
	//get
	for i := 0; i < 10000; i++ {
		valOffical, _ := mapOffical[strconv.Itoa(i)]
		valAny, _ := m.GetString(strconv.Itoa(i))
		valAnyUnsafe, _ := m.GetUnsafe([]byte(strconv.Itoa(i)))
		if valOffical != valAny {
			t.Fatalf("unexpected value obtained; got %q want %q", valAny, valOffical)
		}
		if valOffical != string(valAnyUnsafe) {
			t.Fatalf("unexpected value obtained; got %q want %q", string(valAnyUnsafe), valOffical)
		}
	}
}

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

func TestUint32(t *testing.T) {
	var mapOffical = make(map[string]string)
	var mapAny = NewUint32("mapUint32ForTest")
	//set
	for i := 0; i < 10000; i++ {
		mapOffical[strconv.Itoa(i)] = strconv.Itoa(i)
		mapAny.SetString(uint32(i), strconv.Itoa(i))
	}
	mapAny.SetFinished()
	//get
	for i := 0; i < 10000; i++ {
		valOffical, _ := mapOffical[strconv.Itoa(i)]
		valAny, _ := mapAny.GetString(uint32(i))
		valAnyUnsafe, _ := mapAny.GetUnsafe(uint32(i))
		if valOffical != valAny {
			t.Fatalf("unexpected value obtained; got %q want %q", valAny, valOffical)
		}
		if valOffical != string(valAnyUnsafe) {
			t.Fatalf("unexpected value obtained; got %q want %q", string(valAnyUnsafe), valOffical)
		}
	}
}

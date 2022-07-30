package noGcStaticMap

import (
	"strconv"
	"testing"
)

func TestAny(t *testing.T) {
	var mapOffical = make(map[string]string)
	var mapAny = NewDefault()
	//set
	mapAny.SetString("", "empty")
	mapAny.SetString("empty", "")
	for i := 0; i < 10000; i++ {
		mapOffical[strconv.Itoa(i)] = strconv.Itoa(i)
		mapAny.SetString(strconv.Itoa(i), strconv.Itoa(i))
	}

	mapAny.SetFinished()
	//get
	for i := 0; i < 10000; i++ {
		valOffical, _ := mapOffical[strconv.Itoa(i)]
		valAny, _ := mapAny.GetString(strconv.Itoa(i))
		valAnyUnsafe, _ := mapAny.GetUnsafe([]byte(strconv.Itoa(i)))
		if valOffical != valAny {
			t.Fatalf("unexpected value obtained; got %q want %q", valAny, valOffical)
		}
		if valOffical != string(valAnyUnsafe) {
			t.Fatalf("unexpected value obtained; got %q want %q", string(valAnyUnsafe), valOffical)
		}
	}
	//value empty
	valEmpty, exist := mapAny.GetString("empty")
	if !exist {
		t.Fatalf("unexpected value obtained; got %v want %v", exist, true)
	}
	if valEmpty != "" {
		t.Fatalf("unexpected value obtained; got %q want %q", valEmpty, "")
	}

	//key empty
	valOFKeyEmpty, exist := mapAny.GetString("")
	if !exist {
		t.Fatalf("unexpected value obtained; got %v want %v", exist, true)
	}
	if valOFKeyEmpty != "empty" {
		t.Fatalf("unexpected value obtained; got %q want %q", valOFKeyEmpty, "empty")
	}
	//

}

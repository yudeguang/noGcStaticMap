package noGcStaticMap

import (
	"strconv"
	"testing"
)

func TestInt(t *testing.T) {
	var mapOffical = make(map[string]string)
	var mapAny = NewInt()
	//set
	for i := 0; i < 10000; i++ {
		mapOffical[strconv.Itoa(i)] = strconv.Itoa(i)
		mapAny.SetString(i, strconv.Itoa(i))
	}
	mapAny.SetFinished()
	//get
	for i := 0; i < 10000; i++ {
		valOffical, _ := mapOffical[strconv.Itoa(i)]
		valAny, _ := mapAny.GetString(i)
		valAnyUnsafe, _ := mapAny.GetUnsafe(i)
		if valOffical != valAny {
			t.Fatalf("unexpected value obtained; got %q want %q", valAny, valOffical)
		}
		if valOffical != string(valAnyUnsafe) {
			t.Fatalf("unexpected value obtained; got %q want %q", string(valAnyUnsafe), valOffical)
		}
	}
}

package gson

import "bytes"
import "testing"
import "reflect"

func TestCollateReset(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 1024), 0)
	cltr := config.NewCollate(make([]byte, 1024), 0)

	ref := []interface{}{"sound", "ok", "horn"}
	config.NewValue(ref).Tocollate(clt)
	cltr.Reset(clt.Bytes())
	if value := cltr.Tovalue(); !reflect.DeepEqual(value, ref) {
		t.Errorf("expected %v, got %v", ref, value)
	}
}

func TestCollateEmpty(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 128), 0)
	jsn := config.NewJson(make([]byte, 128), 0)
	clt := config.NewCollate(make([]byte, 128), 0)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		clt.Tovalue()
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		clt.Tojson(jsn)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		clt.Tocbor(cbr)
	}()
}

// sort type for slice of []byte

type ByteSlices [][]byte

func (bs ByteSlices) Len() int {
	return len(bs)
}

func (bs ByteSlices) Less(i, j int) bool {
	return bytes.Compare(bs[i], bs[j]) < 0
}

func (bs ByteSlices) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

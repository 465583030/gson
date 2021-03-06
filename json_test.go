package gson

import "testing"

func TestJsonEmpty(t *testing.T) {
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
		jsn.Tovalue()
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsn.Tocbor(cbr)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsn.Tocollate(clt)
	}()
}

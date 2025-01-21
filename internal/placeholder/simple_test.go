package placeholder

import "testing"

func TestAdd(t *testing.T) {
	if res := Add(1, 2, 3); res != 6 {
		t.Errorf("1+2+3 did not result in 6")
	}
}

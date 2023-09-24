package toolkit

import (
	"testing"
)

func Test_RandomString(t *testing.T) {
	var tool Toolkit
	value := tool.RandomString(10)
	if len(value) != 10 {
		t.Errorf("Expected value of string was 10 but function returned string of length %d", len(value))
	}
}

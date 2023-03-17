package intutil

import "testing"

func Test_Min(t *testing.T) {
	t.Parallel()

	if Min(1, 2) != 1 {
		t.Error("Min(1, 2) != 1")
	}
	if Min(2, 1) != 1 {
		t.Error("Min(2, 1) != 1")
	}
}

func Test_Max(t *testing.T) {
	t.Parallel()

	if Max(1, 2) != 2 {
		t.Error("Max(1, 2) != 2")
	}
	if Max(2, 1) != 2 {
		t.Error("Max(2, 1) != 2")
	}
}

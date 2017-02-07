package driver

import "testing"

func TestDirToDst(t *testing.T) {
	var tests = []struct {
		lastFloor int
		dst       int
		want      string
	}{
		{1, 2, up},
		{3, 3, stop},
		{3, 2, down},
		{0, 0, stop},
	}
	for _, test := range tests {
		if got := dirToDst(test.lastFloor, test.dst); got != test.want {
			t.Errorf("dirToDst(%v,%v) = %s", test.lastFloor, test.dst, got)
		}
	}
}

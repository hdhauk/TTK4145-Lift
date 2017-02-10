package driver

import (
	"fmt"
	"testing"
)

func TestValidateButton(t *testing.T) {
	cfg.Floors = 4
	var tests = []struct {
		b    Btn
		want error
	}{
		{Btn{Floor: 3, Type: HallUp}, fmt.Errorf("no up button at top floor")},
		{Btn{Floor: 0, Type: HallDown}, fmt.Errorf("no down button at ground floor")},
		{Btn{Floor: -3, Type: HallDown}, fmt.Errorf("floor not in range [ %d - %d ]", 0, cfg.Floors-1)},
		{Btn{Floor: 4, Type: HallDown}, fmt.Errorf("floor not in range [ %d - %d ]", 0, cfg.Floors-1)},
		{Btn{Floor: 0, Type: HallUp}, nil},
		{Btn{Floor: 1, Type: HallUp}, nil},
		{Btn{Floor: 2, Type: HallUp}, nil},
		{Btn{Floor: 1, Type: HallDown}, nil},
		{Btn{Floor: 2, Type: HallDown}, nil},
		{Btn{Floor: 3, Type: HallDown}, nil},
	}
	for _, test := range tests {
		got := validateButton(test.b)
		if got == nil {
			if test.want != nil {
				t.Errorf("validateBtn(%+v) == %+v. We want %+v", test.b, got, test.want)
			}
		} else {
			if test.want == nil {
				t.Errorf("validateBtn(%+v) == %+v. We want %+v", test.b, got, test.want)
			}
		}

	}
}

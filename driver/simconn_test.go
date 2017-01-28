package driver

import "testing"

func TestCmdMotorDir(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{"UP", "\x01\x01\x00\x00"},
		{"DOWN", "\x01\xFF\x00\x00"},
		{"STOP", "\x01\x00\x00\x00"},
	}
	for _, test := range tests {
		if got := cmdMotorDir(test.input); got != test.want {
			t.Errorf("cmdMotorDir(%s) = %s", test.input, got)
		}
	}
}

func TestCmdBtnLED(t *testing.T) {
	var tests = []struct {
		inputBtn    Btn
		inputActive bool
		want        string
	}{
		{Btn{0, HallUp}, true, "\x02\x00\x00\x01"},
		{Btn{0, Cab}, true, string([]byte{2, 2, 0, 1})},
		{Btn{1, HallUp}, true, string([]byte{2, 0, 1, 1})},
		{Btn{1, HallDown}, true, string([]byte{2, 1, 1, 1})},
		{Btn{2, HallUp}, false, string([]byte{2, 0, 2, 0})},
		{Btn{2, HallDown}, false, string([]byte{2, 1, 2, 0})},
		{Btn{2, Cab}, true, string([]byte{2, 2, 2, 1})},
	}

	for _, test := range tests {
		got := cmdBtnLED(test.inputBtn, test.inputActive)
		if got != test.want {
			t.Errorf("cmdBtnLED(%v, %v) = %v", test.inputBtn, test.inputActive, []byte(got))
		}
	}
}

func TestCmdFloorLED(t *testing.T) {
	var tests = []struct {
		inputFloor int
		want       string
	}{
		{0, "\x03\x00\x00\x00"},
		{4, "\x03\x04\x00\x00"},
		{255, "\x03\xFF\x00\x00"},
	}

	for _, test := range tests {
		if got := cmdFloorLED(test.inputFloor); got != test.want {
			t.Errorf("cmdFloorLED(%v) = %v", test.inputFloor, []byte(got))
		}
	}
}

func TestCmdDoorLED(t *testing.T) {
	var tests = []struct {
		input bool
		want  string
	}{
		{true, "\x04\x01\x00\x00"},
		{false, "\x04\x00\x00\x00"},
	}

	for _, test := range tests {
		if got := cmdDoorLED(test.input); got != test.want {
			t.Errorf("cmdDoorLED(%t) = %v", test.input, []byte(got))
		}
	}
}

func TestCmdReadOrderBtn(t *testing.T) {
	var testes = []struct {
		input Btn
		want  string
	}{
		{Btn{1, HallUp}, "\x06\x00\x01\x00"},
		{Btn{2, HallDown}, "\x06\x01\x02\x00"},
		{Btn{3, Cab}, "\x06\x02\x03\x00"},
	}

	for _, test := range testes {
		if got := cmdReadOrderBtn(test.input); got != test.want {
			t.Errorf("cmdReadOrderBtn(%v) = %v", test.input, []byte(got))
		}
	}
}

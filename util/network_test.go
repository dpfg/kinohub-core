package util

import "testing"

func TestPadLeft(t *testing.T) {
	type args struct {
		str    string
		pad    string
		lenght int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Prepend 1", args{str: "890", pad: "#", lenght: 4}, "#890"},
		{"Prepend 2", args{str: "90", pad: "#", lenght: 4}, "##90"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadLeft(tt.args.str, tt.args.pad, tt.args.lenght); got != tt.want {
				t.Errorf("PadLeft() = %v, want %v", got, tt.want)
			}
		})
	}
}

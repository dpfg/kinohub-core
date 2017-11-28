package kinopub

import "testing"

func TestParseUID(t *testing.T) {
	type args struct {
		uid string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "", args: args{uid: "KH42"}, want: 42, wantErr: false},
		{name: "", args: args{uid: "TM42"}, want: -1, wantErr: true},
		{name: "", args: args{uid: "42"}, want: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUID(tt.args.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

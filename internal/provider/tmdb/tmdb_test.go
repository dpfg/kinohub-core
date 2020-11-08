package tmdb

import (
	"testing"
)

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
		{name: "", args: args{uid: "KH42"}, want: -1, wantErr: true},
		{name: "", args: args{uid: "TM42"}, want: 42, wantErr: false},
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

func TestImagePath(t *testing.T) {
	type args struct {
		tmdbPath string
		w        int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "", args: args{tmdbPath: "/fjwg4g413", w: -1}, want: "https://image.tmdb.org/t/p/original/fjwg4g413"},
		{name: "", args: args{tmdbPath: "/fjwg4g413", w: 320}, want: "https://image.tmdb.org/t/p/w320/fjwg4g413"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImagePath(tt.args.tmdbPath, tt.args.w); got != tt.want {
				t.Errorf("ImagePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

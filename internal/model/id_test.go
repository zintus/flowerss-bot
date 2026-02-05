package model

import "testing"

func Test_genHashID(t *testing.T) {
	type args struct {
		sLink string
		id    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"case1", args{"http://www.ruanyifeng.com/blog/atom.xml", "tag:www.ruanyifeng.com,2019:/blog//1.2054"},
			"7d35fb910816ebb4",
		},
		{"case2", args{"https://rsshub.app/guokr/scientific", "https://www.guokr.com/article/445877/"}, "561e4b5acbae7324"},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := GenHashID(tt.args.sLink, tt.args.id); got != tt.want {
					t.Errorf("GenHashID() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

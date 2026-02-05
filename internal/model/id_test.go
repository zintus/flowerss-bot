package model

import "testing"

func Test_genHashID(t *testing.T) {
	type args struct {
		sLink   string
		id      string
		rawLink string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"case1",
			args{"http://www.ruanyifeng.com/blog/atom.xml", "tag:www.ruanyifeng.com,2019:/blog//1.2054", ""},
			"7d35fb910816ebb4",
		},
		{
			"case2",
			args{"https://rsshub.app/guokr/scientific", "https://www.guokr.com/article/445877/", ""},
			"561e4b5acbae7324",
		},
		{
			"empty_id_uses_rawLink",
			args{"https://example.com/feed.xml", "", "https://example.com/article/1"},
			"61157afb127305d6",
		},
		{
			"empty_id_different_rawLinks_produce_different_hashes",
			args{"https://example.com/feed.xml", "", "https://example.com/article/2"},
			"61157afb127305d5",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := GenHashID(tt.args.sLink, tt.args.id, tt.args.rawLink); got != tt.want {
					t.Errorf("GenHashID() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_genHashID_emptyID_uniqueness(t *testing.T) {
	// Verify that when id is empty, different rawLinks produce different hashes
	sLink := "https://example.com/feed.xml"
	hash1 := GenHashID(sLink, "", "https://example.com/article/1")
	hash2 := GenHashID(sLink, "", "https://example.com/article/2")

	if hash1 == hash2 {
		t.Errorf("Expected different hashes for different rawLinks when id is empty, got same hash: %v", hash1)
	}
}

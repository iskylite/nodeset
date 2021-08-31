package nodeset

import "testing"

func TestMerge(t *testing.T) {
	type args struct {
		nodestr []string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"L1", args{[]string{"cn0", "cn1"}}},
		{"L2", args{[]string{"cn0,cn2", "cn[1-4,10]"}}},
		{"L3", args{[]string{"cn3", "localhost", "cn0"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Merge(tt.args.nodestr...)
			if err != nil {
				panic(err)
			}
			t.Log(got)
		})
	}
}

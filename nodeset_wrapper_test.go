package nodeset

import (
	"testing"
)

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

func TestExpand(t *testing.T) {
	type args struct {
		nodestr string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"L1", args{"cn[0-2,4-5]"}},
		{"L2", args{"localhost,cn[0-2,4-5]"}},
		{"L3", args{"localhost,cn2ion,cn1ion"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Expand(tt.args.nodestr)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(got)
		})
	}
}

func TestMergeBug(t *testing.T) {
	nodes := []string{"b16r3n00", "b16r3n01", "b16r3n02", "b16r3n03", "b16r3n04", "b16r3n05", "b16r3n06", "b16r3n07", "b16r3n08", "b16r3n09", "b16r3n10", "b16r3n11", "b16r3n12", "b16r3n13", "b16r3n14", "b16r3n15", "b16r3n16", "b16r3n17", "b16r3n18", "b16r3n19", "b19r3n00", "b19r3n01", "b19r3n02", "b19r3n03", "b19r3n04", "b19r3n05", "b19r3n06", "b19r3n07", "b19r3n08", "b19r3n09", "b19r3n10", "b19r3n11", "b19r3n12", "b19r3n13", "b19r3n14", "b19r3n15", "b19r3n16", "b19r3n17", "b19r3n18", "b19r3n19", "b04r4n00", "b04r4n01", "b04r4n02", "b04r4n03", "b04r4n04", "b04r4n05", "b04r4n06", "b04r4n07", "b04r4n08", "b04r4n09", "b04r4n10", "b04r4n11", "b04r4n12", "b04r4n13", "b04r4n14", "b04r4n15", "b04r4n16", "b04r4n17", "b04r4n18", "b19r4n00", "b19r4n01", "b19r4n02", "b19r4n03", "b19r4n04", "b19r4n05", "b19r4n06", "b19r4n07", "b19r4n08", "b19r4n09", "b19r4n10", "b19r4n11", "b19r4n12", "b19r4n13", "b19r4n14", "b19r4n15", "b19r4n16", "b19r4n17", "b19r4n18", "b03r4n00", "b03r4n01", "b03r4n02", "b03r4n03", "b03r4n04", "b03r4n05", "b03r4n06", "b03r4n07", "b03r4n08", "b03r4n09", "b03r4n10", "b03r4n11", "b03r4n12", "b03r4n14", "b03r4n15", "b03r4n16", "b03r4n17", "b03r4n18", "b03r4n19", "b06r4n00", "b06r4n01", "b06r4n02", "b06r4n03", "b06r4n04", "b06r4n05", "b06r4n06", "b06r4n07", "b06r4n08", "b06r4n09", "b06r4n10", "b06r4n11", "b06r4n12", "b06r4n14", "b06r4n15", "b06r4n16", "b06r4n17", "b06r4n18", "b06r4n19", "b15r3n00", "b15r3n01", "b15r3n02", "b15r3n03", "b15r3n04", "b15r3n05", "b15r3n06", "b15r3n07", "b15r3n09", "b15r3n10", "b15r3n11", "b15r3n12", "b15r3n13", "b15r3n14", "b15r3n15", "b15r3n16", "b15r3n17", "b15r3n18", "b15r3n19", "b20r3n00", "b20r3n01", "b20r3n02", "b20r3n03", "b20r3n04", "b20r3n05", "b20r3n06", "b20r3n07", "b20r3n09", "b20r3n10", "b20r3n11", "b20r3n12", "b20r3n13", "b20r3n14", "b20r3n15", "b20r3n16", "b20r3n17", "b20r3n18", "b20r3n19", "b01r3n01", "b01r3n02", "b01r3n03", "b01r3n04", "b01r3n05", "b01r3n06", "b01r3n07", "b01r3n08", "b01r3n09", "b01r3n10", "b01r3n11", "b01r3n12", "b01r3n13", "b01r3n14", "b01r3n15", "b01r3n16", "b01r3n17", "b01r3n18", "b01r3n19", "b01r4n01", "b01r4n02", "b01r4n03", "b01r4n04", "b01r4n05", "b01r4n06", "b01r4n07", "b01r4n08", "b01r4n09", "b01r4n10", "b01r4n11", "b01r4n12", "b01r4n13", "b01r4n14", "b01r4n15", "b01r4n16", "b01r4n17", "b01r4n18", "b01r4n19", "b06r3n00", "b06r3n01", "b06r3n02", "b06r3n03", "b06r3n04", "b06r3n05", "b06r3n07", "b06r3n08", "b06r3n09", "b06r3n10", "b06r3n11", "b06r3n12", "b06r3n13", "b06r3n14", "b06r3n15", "b06r3n17", "b06r3n18", "b06r3n19", "b18r3n00", "b18r3n01", "b18r3n02", "b18r3n03", "b18r3n04", "b18r3n05", "b18r3n07", "b18r3n08", "b18r3n09", "b18r3n10", "b18r3n11", "b18r3n12", "b18r3n13", "b18r3n14", "b18r3n15", "b18r3n17", "b18r3n18", "b18r3n19", "b07r4n00", "b07r4n01", "b07r4n02", "b07r4n03", "b07r4n04", "b07r4n06", "b07r4n07", "b07r4n08", "b07r4n09", "b07r4n10", "b07r4n11", "b07r4n12", "b07r4n13", "b07r4n14", "b07r4n16", "b07r4n17", "b07r4n18", "b07r4n19", "b12r4n00", "b12r4n01", "b12r4n02", "b12r4n03", "b12r4n04", "b12r4n06", "b12r4n07", "b12r4n08", "b12r4n09", "b12r4n10", "b12r4n11", "b12r4n12", "b12r4n13", "b12r4n14", "b12r4n16", "b12r4n17", "b12r4n18", "b12r4n19", "b19r1n00", "b19r1n01", "b19r1n02", "b19r1n03", "b19r1n04", "b19r1n05", "b19r1n06", "b19r1n07", "b19r1n08", "b19r1n09", "b19r1n10", "b19r1n11", "b19r1n12", "b19r1n13", "b19r1n14", "b19r1n15", "b19r1n16", "b19r1n17", "b19r1n18"}
	node := nodes[0]
	nodesNum := len(nodes)
	nodelist, err := Merge(nodes[1:nodesNum]...)
	if err != nil {
		t.Log(err)
	}
	t.Logf("[%s] -> %s\n", node, nodelist)
}

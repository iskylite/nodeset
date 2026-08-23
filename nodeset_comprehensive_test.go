package nodeset

import (
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func TestComplexExpressionsExpandFoldAndRoundTrip(t *testing.T) {
	tests := []struct {
		expr       string
		wantLen    int
		wantFirst  string
		wantLast   string
		wantString string
	}{
		{"node[0-100]", 101, "node0", "node100", "node[0-100]"},
		{"node[001-100]", 100, "node001", "node100", "node[001-100]"},
		{"mn[0-10],cn[0-10]", 22, "cn0", "mn10", "cn[0-10],mn[0-10]"},
		{"a[01,10]r[1-4]n[01-20]", 160, "a01r1n01", "a10r4n20", "a[01,10]r[1-4]n[01-20]"},
		{"10.0.0.[1-3]", 3, "10.0.0.1", "10.0.0.3", "10.0.0.[1-3]"},
		{"2001:db8::[1-3]", 3, "2001:db8::1", "2001:db8::3", "2001:db8::[1-3]"},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			ns, err := NewNodeSet(tt.expr)
			if err != nil {
				t.Fatal(err)
			}
			got := expandNodeSet(ns)
			if len(got) != tt.wantLen {
				t.Fatalf("len=%d want %d; nodes=%v", len(got), tt.wantLen, got)
			}
			if got[0] != tt.wantFirst || got[len(got)-1] != tt.wantLast {
				t.Fatalf("boundary=(%q,%q) want (%q,%q)", got[0], got[len(got)-1], tt.wantFirst, tt.wantLast)
			}
			if got := ns.String(); got != tt.wantString {
				t.Fatalf("String()=%q want %q", got, tt.wantString)
			}
			roundTrip, err := NewNodeSet(ns.String())
			if err != nil {
				t.Fatal(err)
			}
			if got := expandNodeSet(roundTrip); !reflect.DeepEqual(got, expandNodeSet(ns)) {
				t.Fatalf("round trip changed nodes: %v -> %v", expandNodeSet(ns), got)
			}
		})
	}
}

func TestNodeSetRejectsEmptyAndMalformedMembers(t *testing.T) {
	for _, expr := range []string{
		"node[1-2],",
		",node[1-2]",
		"node[1-2],,node[3-4]",
		"node[]",
		"node[1-2]tail]",
		"node[1-2tail",
		"node[1-3/0]",
		"node[1/2]",
		"foo%s",
		"foo%d[1-2]",
		"192.168.1.0/33",
		"2001:db8::/129",
		"10.0.0.0/0",
	} {
		t.Run(expr, func(t *testing.T) {
			if _, err := NewNodeSet(expr); err == nil {
				t.Fatalf("NewNodeSet(%q) unexpectedly succeeded", expr)
			}
		})
	}
}

func TestCIDRExpansionAndMixedExpressions(t *testing.T) {
	tests := []struct {
		expr       string
		want       []string
		wantString string
	}{
		{
			expr:       "192.168.1.0/30",
			want:       []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3"},
			wantString: "192.168.1.[0-3]",
		},
		{
			expr:       "192.168.1.1/30",
			want:       []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3"},
			wantString: "192.168.1.[0-3]",
		},
		{
			expr:       "2001:db8::2/126",
			want:       []string{"2001:db8::", "2001:db8::1", "2001:db8::2", "2001:db8::3"},
			wantString: "2001:db8::,2001:db8::[1-3]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := Expand(tt.expr)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Expand=%v want %v", got, tt.want)
			}
			ns, err := NewNodeSet(tt.expr)
			if err != nil {
				t.Fatal(err)
			}
			if got := ns.String(); got != tt.wantString {
				t.Fatalf("String=%q want %q", got, tt.wantString)
			}
		})
	}

	got, err := Expand("node[1-2],192.168.1.0/30")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3", "node1", "node2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mixed Expand=%v want %v", got, want)
	}
}

func TestNodeSetAddIsTransactional(t *testing.T) {
	ns, err := NewNodeSet("base")
	if err != nil {
		t.Fatal(err)
	}
	if err := ns.Add("good,node["); err == nil {
		t.Fatal("Add unexpectedly succeeded")
	}
	if got := ns.String(); got != "base" {
		t.Fatalf("failed Add partially mutated set: %q", got)
	}
}

func TestNodeSetStringAndIteratorAreDeterministic(t *testing.T) {
	const expr = "z[0-10],a[0-10],mn[0-10],cn[0-10],localhost"
	ns, err := NewNodeSet(expr)
	if err != nil {
		t.Fatal(err)
	}
	want := ns.String()
	for i := 0; i < 100; i++ {
		if got := ns.String(); got != want {
			t.Fatalf("String() changed between calls: %q then %q", want, got)
		}
		if got := expandNodeSet(ns); !reflect.DeepEqual(got, expandNodeSet(ns)) {
			t.Fatalf("Iterator order changed: %v", got)
		}
	}
}

func TestSplitRejectsNonPositiveWidth(t *testing.T) {
	if got := Split([]string{"n1", "n2"}, 0); len(got) != 0 {
		t.Fatalf("Split width 0=%v want empty", got)
	}
	if got := Split([]string{"n1", "n2"}, -1); len(got) != 0 {
		t.Fatalf("Split width -1=%v want empty", got)
	}
}

func TestExpandSortsAllPatternsNaturally(t *testing.T) {
	ns, err := NewNodeSet("node[10-12],node[2-4],node1")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"node1", "node2", "node3", "node4", "node10", "node11", "node12"}
	if got := expandNodeSet(ns); !reflect.DeepEqual(got, want) {
		t.Fatalf("Expand=%v want %v", got, want)
	}
	if got := ns.String(); got != "node[1-4,10-12]" {
		t.Fatalf("String=%q", got)
	}
}

func TestMergeRoundTripRandomized(t *testing.T) {
	rng := rand.New(rand.NewSource(20260823))
	for iteration := 0; iteration < 100; iteration++ {
		seen := make(map[string]struct{})
		for len(seen) < 1+rng.Intn(120) {
			a := rng.Intn(8)
			r := 1 + rng.Intn(4)
			n := rng.Intn(20)
			seen["a"+strconv.Itoa(a)+"r"+strconv.Itoa(r)+"n"+formatTwoDigits(n)] = struct{}{}
		}
		input := make([]string, 0, len(seen))
		for node := range seen {
			input = append(input, node)
		}
		merged, err := Merge(input...)
		if err != nil {
			t.Fatalf("iteration %d Merge error: %v", iteration, err)
		}
		roundTrip, err := NewNodeSet(merged)
		if err != nil {
			t.Fatalf("iteration %d round-trip parse %q: %v", iteration, merged, err)
		}
		want := sortedStrings(input)
		got := sortedStrings(expandNodeSet(roundTrip))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("iteration %d changed set: merged=%q got=%v want=%v", iteration, merged, got, want)
		}
	}
}

func expandNodeSet(ns *NodeSet) []string {
	it := ns.Iterator()
	result := make([]string, 0, it.Len())
	for it.Next() {
		result = append(result, it.Value())
	}
	return result
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func formatTwoDigits(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}

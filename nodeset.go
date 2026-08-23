package nodeset

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	ErrInvalidNodeSet = errors.New("invalid nodeset")
	ErrParseNodeSet   = errors.New("nodeset parse error")
	rangeSetRegexp    = regexp.MustCompile(`(\[[^\[\]]+\]|[0-9]+)`)
)

type Pattern struct {
	format   string
	rangeSet *RangeSetND
}

type NodeSet struct {
	patterns map[string]*RangeSetND
}

func EmptyNodeSet() *NodeSet {
	return &NodeSet{patterns: make(map[string]*RangeSetND)}
}

func NewNodeSet(nodestr string) (*NodeSet, error) {
	ns := &NodeSet{patterns: make(map[string]*RangeSetND)}

	if nodestr == "" {
		// Empty nodeset
		return ns, nil
	}

	err := ns.Add(nodestr)
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func (ns *NodeSet) Add(nodestr string) error {
	if ns == nil {
		return fmt.Errorf("nil nodeset - %w", ErrInvalidNodeSet)
	}
	parsed, err := parseNodeSet(nodestr)
	if err != nil {
		return err
	}
	if ns.patterns == nil {
		ns.patterns = make(map[string]*RangeSetND)
	}
	for pattern, rs := range parsed.patterns {
		if current, ok := ns.patterns[pattern]; ok && current != nil && rs != nil && current.Dim() != rs.Dim() {
			return fmt.Errorf("mismatched dimensions %d != %d - %w", current.Dim(), rs.Dim(), ErrInvalidNodeSet)
		}
	}
	for pattern, rs := range parsed.patterns {
		if current, ok := ns.patterns[pattern]; ok && current != nil && rs != nil {
			if err := current.Update(rs); err != nil {
				return err
			}
			continue
		}
		ns.patterns[pattern] = rs
	}
	return nil
}

func parseNodeSet(nodestr string) (*NodeSet, error) {
	members, err := splitExpressionMembers(nodestr)
	if err != nil {
		return nil, err
	}
	var expanded []string
	for _, member := range members {
		member = strings.TrimSpace(member)
		if member == "" {
			return nil, fmt.Errorf("empty node member in %s - %w", nodestr, ErrParseNodeSet)
		}
		if topLevelSlash(member) {
			ips, err := ExpandCIDR(member)
			if err != nil {
				return nil, err
			}
			expanded = append(expanded, ips...)
		} else {
			expanded = append(expanded, member)
		}
	}
	nodestr = strings.Join(expanded, ",")
	ns := &NodeSet{patterns: make(map[string]*RangeSetND)}
	nodestr = strings.ReplaceAll(nodestr, " ", "")
	if nodestr == "" {
		return nil, fmt.Errorf("empty nodeset - %w", ErrParseNodeSet)
	}
	if strings.Contains(nodestr, "%") {
		return nil, fmt.Errorf("invalid format character in %s - %w", nodestr, ErrParseNodeSet)
	}

	ranges := rangeSetRegexp.FindAllStringSubmatch(nodestr, -1)

	patterns := rangeSetRegexp.ReplaceAllString(nodestr, "%s")

	if strings.Contains(patterns, "[") {
		return nil, fmt.Errorf("unbalanced '[' found while parsing %s - %w", nodestr, ErrParseNodeSet)
	}
	if strings.Contains(patterns, "]") {
		return nil, fmt.Errorf("unbalanced ']' found while parsing %s - %w", nodestr, ErrParseNodeSet)
	}

	ridx := 0
	for _, pattern := range strings.Split(patterns, ",") {
		if pattern == "" {
			return nil, fmt.Errorf("empty node member in %s - %w", nodestr, ErrParseNodeSet)
		}
		rangeSetCount := strings.Count(pattern, "%s")
		if rangeSetCount == 0 {
			ns.patterns[pattern] = nil
			continue
		}

		rangeSets := make([]string, 0)
		for i := ridx; i < ridx+rangeSetCount; i++ {
			rangeSets = append(rangeSets, strings.Trim(ranges[i][1], "[]"))
		}
		rs, err := NewRangeSetND([][]string{rangeSets})
		if err != nil {
			return nil, err
		}

		if _, ok := ns.patterns[pattern]; !ok {
			ns.patterns[pattern] = rs
		} else {
			err = ns.patterns[pattern].Update(rs)
			if err != nil {
				return nil, err
			}
		}
		ridx += rangeSetCount
	}

	return ns, nil
}

func splitExpressionMembers(expr string) ([]string, error) {
	depth := 0
	start := 0
	result := make([]string, 0, 1)
	for i, r := range expr {
		switch r {
		case '[':
			depth++
		case ']':
			if depth == 0 {
				return nil, fmt.Errorf("unbalanced ']' found while parsing %s - %w", expr, ErrParseNodeSet)
			}
			depth--
		case ',':
			if depth == 0 {
				result = append(result, expr[start:i])
				start = i + 1
			}
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("unbalanced '[' found while parsing %s - %w", expr, ErrParseNodeSet)
	}
	return append(result, expr[start:]), nil
}

func topLevelSlash(expr string) bool {
	depth := 0
	for _, r := range expr {
		switch r {
		case '[':
			depth++
		case ']':
			depth--
		case '/':
			if depth == 0 {
				return true
			}
		}
	}
	return false
}

func (ns *NodeSet) Len() int {
	size := 0

	for _, rs := range ns.patterns {
		if rs == nil {
			size++
		} else {
			size += rs.Len()
		}
	}

	return size
}

func (ns *NodeSet) String() string {
	list := ns.toStringList()
	return strings.Join(list, ",")
}

func (ns *NodeSet) toStringList() []string {
	list := make([]string, 0)

	items := make([]*Pattern, 0, len(ns.patterns))
	for pattern, rs := range ns.patterns {
		items = append(items, &Pattern{format: pattern, rangeSet: rs})
	}

	for _, pattern := range items {
		if pattern.rangeSet == nil {
			list = append(list, pattern.format)
			continue
		}
		for _, params := range pattern.rangeSet.FormatList() {
			list = append(list, fmt.Sprintf(pattern.format, params...))
		}

	}
	sort.Slice(list, func(i, j int) bool { return naturalLess(list[i], list[j]) })
	return list
}

func (ns *NodeSet) MarshalJSON() ([]byte, error) {
	list := ns.toStringList()
	return json.Marshal(&list)
}

func (ns *NodeSet) UnmarshalJSON(data []byte) error {
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	n, err := NewNodeSet(strings.Join(list, ","))
	if err != nil {
		return err
	}

	*ns = *n

	return nil
}

func (ns *NodeSet) Iterator() *NodeSetIterator {
	nodes := make([]string, 0, ns.Len())

	items := make([]*Pattern, 0, len(ns.patterns))
	for pattern, rs := range ns.patterns {
		items = append(items, &Pattern{format: pattern, rangeSet: rs})
	}

	for _, pattern := range items {
		if pattern.rangeSet == nil {
			nodes = append(nodes, pattern.format)
			continue
		}

		it := pattern.rangeSet.Iterator()
		for it.Next() {
			params := it.FormatList()
			nodes = append(nodes, fmt.Sprintf(pattern.format, params...))
		}

	}
	sort.Slice(nodes, func(i, j int) bool { return naturalLess(nodes[i], nodes[j]) })

	return &NodeSetIterator{nodes: nodes, current: -1}
}

func naturalLess(a, b string) bool {
	for len(a) > 0 && len(b) > 0 {
		adigit := a[0] >= '0' && a[0] <= '9'
		bdigit := b[0] >= '0' && b[0] <= '9'
		if adigit && bdigit {
			aEnd := 0
			for aEnd < len(a) && a[aEnd] >= '0' && a[aEnd] <= '9' {
				aEnd++
			}
			bEnd := 0
			for bEnd < len(b) && b[bEnd] >= '0' && b[bEnd] <= '9' {
				bEnd++
			}
			an := strings.TrimLeft(a[:aEnd], "0")
			bn := strings.TrimLeft(b[:bEnd], "0")
			if an == "" {
				an = "0"
			}
			if bn == "" {
				bn = "0"
			}
			if len(an) != len(bn) {
				return len(an) < len(bn)
			}
			if an != bn {
				return an < bn
			}
			if a[:aEnd] != b[:bEnd] {
				return len(a[:aEnd]) < len(b[:bEnd])
			}
			a, b = a[aEnd:], b[bEnd:]
			continue
		}
		if a[0] != b[0] {
			return a[0] < b[0]
		}
		a, b = a[1:], b[1:]
	}
	return len(a) < len(b)
}

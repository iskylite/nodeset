package nodeset

import (
	"errors"
	"fmt"
	"net"
)

const MaxExpand = 1_000_000

var (
	ErrInvalidCIDR = errors.New("invalid CIDR")
	ErrExpandLimit = errors.New("CIDR expansion exceeds limit")
)

// ExpandCIDR expands one IPv4 or IPv6 network. Host bits in the input are
// ignored according to net.ParseCIDR semantics; expansion always starts at
// the network address.
func ExpandCIDR(cidr string) ([]string, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil || network == nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidCIDR, cidr)
	}
	prefix, width := network.Mask.Size()
	count, ok := cidrExpansionSize(prefix, width)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrExpandLimit, cidr)
	}
	base := append(net.IP(nil), network.IP...)
	if v4 := network.IP.To4(); v4 != nil {
		base = append(net.IP(nil), v4...)
	}
	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		current := append(net.IP(nil), base...)
		incrementIP(current, i)
		result = append(result, current.String())
	}
	return result, nil
}

func cidrExpansionSize(prefix, width int) (int, bool) {
	if prefix < 0 || width <= 0 || prefix > width {
		return 0, false
	}
	hostBits := width - prefix
	if hostBits >= 63 {
		return 0, false
	}
	count := uint64(1) << uint(hostBits)
	if count > MaxExpand {
		return 0, false
	}
	return int(count), true
}

func incrementIP(ip net.IP, amount int) {
	for i := len(ip) - 1; i >= 0 && amount > 0; i-- {
		amount += int(ip[i])
		ip[i] = byte(amount)
		amount >>= 8
	}
}

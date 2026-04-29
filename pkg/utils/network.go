package utils

import (
	"fmt"
	"net/netip"
)

func ParseIpToPrefix(s string) (netip.Prefix, error) {
	if p, err := netip.ParsePrefix(s); err == nil {
		return p, nil
	}

	if ip, err := netip.ParseAddr(s); err == nil {
		return netip.PrefixFrom(ip, ip.BitLen()), nil
	}

	return netip.Prefix{}, fmt.Errorf("invalid ip or cidr: %s", s)
}

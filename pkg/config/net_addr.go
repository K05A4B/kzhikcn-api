package config

import (
	"fmt"
	"net/netip"

	"gopkg.in/yaml.v3"
)

type NetAddr netip.Prefix

func (n *NetAddr) UnmarshalYAML(value *yaml.Node) error {
	str := ""
	err := value.Decode(&str)
	if err != nil {
		return err
	}

	prefix, err := netip.ParsePrefix(str)

	if err == nil {
		*n = NetAddr(prefix)
		return nil
	}

	ip, err := netip.ParseAddr(str)
	if err == nil {
		*n = NetAddr(netip.PrefixFrom(ip, ip.BitLen()))
		return nil
	}

	return fmt.Errorf("invalid ip or cidr: %s", str)
}

func (n *NetAddr) ToPrefix() netip.Prefix {
	return netip.Prefix(*n)
}

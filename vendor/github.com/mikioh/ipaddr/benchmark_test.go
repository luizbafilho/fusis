// Copyright 2015 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"net"
	"testing"

	"github.com/mikioh/ipaddr"
)

var (
	aggregatablePrefixesIPv4 = toPrefixes([]string{
		"192.0.2.0/28", "192.0.2.16/28", "192.0.2.32/28", "192.0.2.48/28",
		"192.0.2.64/28", "192.0.2.80/28", "192.0.2.96/28", "192.0.2.112/28",
		"192.0.2.128/28", "192.0.2.144/28", "192.0.2.160/28", "192.0.2.176/28",
		"192.0.2.192/28", "192.0.2.208/28", "192.0.2.224/28", "192.0.2.240/28",
		"198.51.100.0/24", "203.0.113.0/24",
	})
	aggregatablePrefixesIPv6 = toPrefixes([]string{
		"2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64",
		"2001:db8:0:4::/64", "2001:db8:0:5::/64", "2001:db8:0:6::/64", "2001:db8:0:7::/64",
		"2001:db8:cafe::/64", "2001:db8:babe::/64",
	})
	ipv4Pair = []*ipaddr.Prefix{toPrefix("192.0.2.0/25"), toPrefix("192.0.2.128/25")}
	ipv6Pair = []*ipaddr.Prefix{toPrefix("2001:db8:f001:f002::/64"), toPrefix("2001:db8:f001:f003::/64")}
)

func BenchmarkAggregateIPv4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipaddr.Aggregate(aggregatablePrefixesIPv4)
	}
}

func BenchmarkAggregateIPv6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipaddr.Aggregate(aggregatablePrefixesIPv6)
	}
}

func BenchmarkCompareIPv4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipaddr.Compare(ipv4Pair[0], ipv4Pair[1])
	}
}

func BenchmarkCompareIPv6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipaddr.Compare(ipv6Pair[0], ipv6Pair[1])
	}
}

func BenchmarkSummarizeIPv4(b *testing.B) {
	fip, lip := net.IPv4(192, 0, 2, 1), net.IPv4(192, 0, 2, 255)

	for i := 0; i < b.N; i++ {
		ipaddr.Summarize(fip, lip)
	}
}

func BenchmarkSummarizeIPv6(b *testing.B) {
	fip, lip := net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::00ff")

	for i := 0; i < b.N; i++ {
		ipaddr.Summarize(fip, lip)
	}
}

func BenchmarkSupernetIPv4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipaddr.Supernet(aggregatablePrefixesIPv4)
	}
}

func BenchmarkSupernetIPv6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipaddr.Supernet(aggregatablePrefixesIPv6)
	}
}

func BenchmarkCursorNextIPv4(b *testing.B) {
	ps := toPrefixes([]string{"192.0.2.0/24"})
	c := ipaddr.NewCursor(ps)

	for i := 0; i < b.N; i++ {
		for c.Next() != nil {
		}
	}
}

func BenchmarkCursorNextIPv6(b *testing.B) {
	ps := toPrefixes([]string{"2001:db8::/120"})
	c := ipaddr.NewCursor(ps)

	for i := 0; i < b.N; i++ {
		for c.Next() != nil {
		}
	}
}

func BenchmarkCursorPrevIPv4(b *testing.B) {
	ps := toPrefixes([]string{"192.0.2.255/24"})
	c := ipaddr.NewCursor(ps)

	for i := 0; i < b.N; i++ {
		for c.Prev() != nil {
		}
	}
}

func BenchmarkCursorPrevIPv6(b *testing.B) {
	ps := toPrefixes([]string{"2001:db8::ff/120"})
	c := ipaddr.NewCursor(ps)

	for i := 0; i < b.N; i++ {
		for c.Prev() != nil {
		}
	}
}

func BenchmarkPrefixEqualIPv4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipv4Pair[0].Equal(ipv4Pair[1])
	}
}

func BenchmarkPrefixEqualIPv6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipv6Pair[0].Equal(ipv6Pair[1])
	}
}

func BenchmarkPrefixExcludeIPv4(b *testing.B) {
	p1, p2 := toPrefix("192.0.2.0/24"), toPrefix("192.0.2.192/32")

	for i := 0; i < b.N; i++ {
		p1.Exclude(p2)
	}
}

func BenchmarkPrefixExcludeIPv6(b *testing.B) {
	p1, p2 := toPrefix("2001:db8::/120"), toPrefix("2001:db8::1/128")

	for i := 0; i < b.N; i++ {
		p1.Exclude(p2)
	}
}

func BenchmarkPrefixMarshalBinaryIPv4(b *testing.B) {
	p := toPrefix("192.0.2.0/31")

	for i := 0; i < b.N; i++ {
		p.MarshalBinary()
	}
}

func BenchmarkPrefixMarshalBinaryIPv6(b *testing.B) {
	p := toPrefix("2001:db8:cafe:babe::/127")

	for i := 0; i < b.N; i++ {
		p.MarshalBinary()
	}
}

func BenchmarkPrefixMarshalTextIPv4(b *testing.B) {
	p := toPrefix("192.0.2.0/31")

	for i := 0; i < b.N; i++ {
		p.MarshalText()
	}
}

func BenchmarkPrefixMarshalTextIPv6(b *testing.B) {
	p := toPrefix("2001:db8:cafe:babe::/127")

	for i := 0; i < b.N; i++ {
		p.MarshalText()
	}
}

func BenchmarkPrefixOverlapsIPv4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipv4Pair[0].Overlaps(ipv4Pair[1])
	}
}

func BenchmarkPrefixOverlapsIPv6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipv6Pair[0].Overlaps(ipv6Pair[1])
	}
}

func BenchmarkPrefixSubnetsIPv4(b *testing.B) {
	p := toPrefix("192.0.2.0/24")

	for i := 0; i < b.N; i++ {
		p.Subnets(3)
	}
}

func BenchmarkPrefixSubnetsIPv6(b *testing.B) {
	p := toPrefix("2001:db8::/32")

	for i := 0; i < b.N; i++ {
		p.Subnets(3)
	}
}

func BenchmarkPrefixUnmarshalBinaryIPv4(b *testing.B) {
	p := toPrefix("0.0.0.0/0")

	for i := 0; i < b.N; i++ {
		p.UnmarshalBinary([]byte{22, 192, 168, 0})
	}
}

func BenchmarkPrefixUnmarshalBinaryIPv6(b *testing.B) {
	p := toPrefix("::/0")

	for i := 0; i < b.N; i++ {
		p.UnmarshalBinary([]byte{66, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0xca, 0xfe, 0x80})
	}
}

func BenchmarkPrefixUnmarshalTextIPv4(b *testing.B) {
	p := toPrefix("0.0.0.0/0")

	for i := 0; i < b.N; i++ {
		p.UnmarshalText([]byte("192.168.0.0/31"))
	}
}

func BenchmarkPrefixUnmarshalTextIPv6(b *testing.B) {
	p := toPrefix("::/0")

	for i := 0; i < b.N; i++ {
		p.UnmarshalText([]byte("2001:db8:cafe:babe::/127"))
	}
}

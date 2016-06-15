// Copyright 2015 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"net"
	"testing"

	"github.com/mikioh/ipaddr"
)

func BenchmarkAggregateIPv4(b *testing.B) {
	in := toPrefixes([]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24", "192.168.3.0/24", "192.168.4.0/25", "192.168.101.0/26", "192.168.102.1/27"})

	for i := 0; i < b.N; i++ {
		ipaddr.Aggregate(in)
	}
}

func BenchmarkAggregateIPv6(b *testing.B) {
	in := toPrefixes([]string{"2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:4::/64", "2001:db8::/64", "2001:db8::1/64"})

	for i := 0; i < b.N; i++ {
		ipaddr.Aggregate(in)
	}
}

func BenchmarkCompareIPv4(b *testing.B) {
	p1 := toPrefix("192.168.1.0/24")
	p2 := toPrefix("192.168.2.0/24")

	for i := 0; i < b.N; i++ {
		ipaddr.Compare(p1, p2)
	}
}

func BenchmarkCompareIPv6(b *testing.B) {
	p1 := toPrefix("2001:db8:f001:f002::/64")
	p2 := toPrefix("2001:db8:f001:f003::/64")

	for i := 0; i < b.N; i++ {
		ipaddr.Compare(p1, p2)
	}
}

func BenchmarkSummarizeIPv4(b *testing.B) {
	fip, lip := net.IPv4(172, 16, 1, 1), net.IPv4(172, 16, 255, 255)

	for i := 0; i < b.N; i++ {
		ipaddr.Summarize(fip, lip)
	}
}

func BenchmarkSummarizeIPv6(b *testing.B) {
	fip, lip := net.ParseIP("2001:db8::1:1"), net.ParseIP("2001:db8::1:ffff")

	for i := 0; i < b.N; i++ {
		ipaddr.Summarize(fip, lip)
	}
}

func BenchmarkSupernetIPv4(b *testing.B) {
	in := toPrefixes([]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24", "192.168.3.0/24", "192.168.4.0/25", "192.168.101.0/26", "192.168.102.1/27"})

	for i := 0; i < b.N; i++ {
		ipaddr.Supernet(in)
	}
}

func BenchmarkSupernetIPv6(b *testing.B) {
	in := toPrefixes([]string{"2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:4::/64", "2001:db8::/64", "2001:db8::1/64"})

	for i := 0; i < b.N; i++ {
		ipaddr.Supernet(in)
	}
}

func BenchmarkCursorNextIPv4(b *testing.B) {
	ps := toPrefixes([]string{"192.168.0.0/24"})
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
	ps := toPrefixes([]string{"192.168.0.255/24"})
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
	ps := toPrefixes([]string{"192.168.1.0/24", "192.168.2.0/24"})

	for i := 0; i < b.N; i++ {
		ps[0].Equal(&ps[1])
	}
}

func BenchmarkPrefixEqualIPv6(b *testing.B) {
	ps := toPrefixes([]string{"2001:db8:f001:f002::/64", "2001:db8:f001:f003::/64"})

	for i := 0; i < b.N; i++ {
		ps[0].Equal(&ps[1])
	}
}

func BenchmarkPrefixExcludeIPv4(b *testing.B) {
	p1 := toPrefix("10.1.0.0/16")
	p2 := toPrefix("10.1.1.1/32")

	for i := 0; i < b.N; i++ {
		p1.Exclude(p2)
	}
}

func BenchmarkPrefixExcludeIPv6(b *testing.B) {
	p1 := toPrefix("2001:db8::/64")
	p2 := toPrefix("2001:db8::1:1:1:1/128")

	for i := 0; i < b.N; i++ {
		p1.Exclude(p2)
	}
}

func BenchmarkPrefixMarshalBinaryIPv4(b *testing.B) {
	p := toPrefix("192.168.0.0/31")

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
	p := toPrefix("192.168.0.0/31")

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
	p1 := toPrefix("192.168.1.0/24")
	p2 := toPrefix("192.168.2.0/24")

	for i := 0; i < b.N; i++ {
		p1.Overlaps(p2)
	}
}

func BenchmarkPrefixOverlapsIPv6(b *testing.B) {
	p1 := toPrefix("2001:db8:f001:f002::/64")
	p2 := toPrefix("2001:db8:f001:f003::/64")

	for i := 0; i < b.N; i++ {
		p1.Overlaps(p2)
	}
}

func BenchmarkPrefixSubnetsIPv4(b *testing.B) {
	p := toPrefix("192.168.0.0/16")

	for i := 0; i < b.N; i++ {
		p.Subnets(3)
	}
}

func BenchmarkPrefixSubnetsIPv6(b *testing.B) {
	p := toPrefix("2001:db8::/60")

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

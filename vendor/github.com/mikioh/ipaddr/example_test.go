// Copyright 2015 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"fmt"
	"log"
	"net"

	"github.com/mikioh/ipaddr"
)

func ExampleCursor_traversal() {
	c, err := ipaddr.Parse("2001:db8::/126,192.168.1.128/30,192.168.0.0/29")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(c.Pos(), c.First(), c.Last(), c.List())
	for pos := c.Next(); pos != nil; pos = c.Next() {
		fmt.Println(pos)
	}
	fmt.Println(c.Pos(), c.First(), c.Last(), c.List())
	for pos := c.Prev(); pos != nil; pos = c.Prev() {
		fmt.Println(pos)
	}
	fmt.Println(c.Pos(), c.First(), c.Last(), c.List())
	// Output:
	// &{192.168.0.0 192.168.0.0/29} &{192.168.0.0 192.168.0.0/29} &{2001:db8::3 2001:db8::/126} [192.168.0.0/29 192.168.1.128/30 2001:db8::/126]
	// &{192.168.0.1 192.168.0.0/29}
	// &{192.168.0.2 192.168.0.0/29}
	// &{192.168.0.3 192.168.0.0/29}
	// &{192.168.0.4 192.168.0.0/29}
	// &{192.168.0.5 192.168.0.0/29}
	// &{192.168.0.6 192.168.0.0/29}
	// &{192.168.0.7 192.168.0.0/29}
	// &{192.168.1.128 192.168.1.128/30}
	// &{192.168.1.129 192.168.1.128/30}
	// &{192.168.1.130 192.168.1.128/30}
	// &{192.168.1.131 192.168.1.128/30}
	// &{2001:db8:: 2001:db8::/126}
	// &{2001:db8::1 2001:db8::/126}
	// &{2001:db8::2 2001:db8::/126}
	// &{2001:db8::3 2001:db8::/126}
	// &{2001:db8::3 2001:db8::/126} &{192.168.0.0 192.168.0.0/29} &{2001:db8::3 2001:db8::/126} [192.168.0.0/29 192.168.1.128/30 2001:db8::/126]
	// &{2001:db8::2 2001:db8::/126}
	// &{2001:db8::1 2001:db8::/126}
	// &{2001:db8:: 2001:db8::/126}
	// &{192.168.1.131 192.168.1.128/30}
	// &{192.168.1.130 192.168.1.128/30}
	// &{192.168.1.129 192.168.1.128/30}
	// &{192.168.1.128 192.168.1.128/30}
	// &{192.168.0.7 192.168.0.0/29}
	// &{192.168.0.6 192.168.0.0/29}
	// &{192.168.0.5 192.168.0.0/29}
	// &{192.168.0.4 192.168.0.0/29}
	// &{192.168.0.3 192.168.0.0/29}
	// &{192.168.0.2 192.168.0.0/29}
	// &{192.168.0.1 192.168.0.0/29}
	// &{192.168.0.0 192.168.0.0/29}
	// &{192.168.0.0 192.168.0.0/29} &{192.168.0.0 192.168.0.0/29} &{2001:db8::3 2001:db8::/126} [192.168.0.0/29 192.168.1.128/30 2001:db8::/126]
}

func ExamplePrefix_subnettingAndSupernetting() {
	_, n, err := net.ParseCIDR("172.16.0.0/16")
	if err != nil {
		log.Fatal(err)
	}
	p := ipaddr.NewPrefix(n)
	fmt.Println(p.IP, p.Last(), p.Len(), p.Mask, p.Hostmask())
	fmt.Println()
	ps := p.Subnets(3)
	for _, p := range ps {
		fmt.Println(p)
	}
	fmt.Println()
	fmt.Println(ipaddr.Supernet(ps))
	fmt.Println(ipaddr.Supernet(ps[:2]))
	fmt.Println(ipaddr.Supernet(ps[2:4]))
	fmt.Println(ipaddr.Supernet(ps[4:6]))
	fmt.Println(ipaddr.Supernet(ps[6:8]))
	// Output:
	// 172.16.0.0 172.16.255.255 16 ffff0000 0000ffff
	//
	// 172.16.0.0/19
	// 172.16.32.0/19
	// 172.16.64.0/19
	// 172.16.96.0/19
	// 172.16.128.0/19
	// 172.16.160.0/19
	// 172.16.192.0/19
	// 172.16.224.0/19
	//
	// 172.16.0.0/16
	// 172.16.0.0/18
	// 172.16.64.0/18
	// 172.16.128.0/18
	// 172.16.192.0/18
}

func ExamplePrefix_subnettingAndAggregation() {
	_, n, err := net.ParseCIDR("192.168.0.0/24")
	if err != nil {
		log.Fatal(err)
	}
	p := ipaddr.NewPrefix(n)
	fmt.Println(p.IP, p.Last(), p.Len(), p.Mask, p.Hostmask())
	fmt.Println()
	ps := p.Subnets(2)
	for _, p := range ps {
		fmt.Println(p)
	}
	fmt.Println()
	_, n, err = net.ParseCIDR("192.168.100.0/24")
	if err != nil {
		log.Fatal(err)
	}
	ps = append(ps, *ipaddr.NewPrefix(n))
	fmt.Println(ipaddr.Aggregate(ps))
	fmt.Println(ipaddr.Aggregate(ps[:2]))
	fmt.Println(ipaddr.Aggregate(ps[2:4]))
	// Output:
	// 192.168.0.0 192.168.0.255 24 ffffff00 000000ff
	//
	// 192.168.0.0/26
	// 192.168.0.64/26
	// 192.168.0.128/26
	// 192.168.0.192/26
	//
	// [192.168.0.0/24 192.168.100.0/24]
	// [192.168.0.0/25]
	// [192.168.0.128/25]
}

func ExamplePrefix_addressRangeSummarization() {
	ps := ipaddr.Summarize(net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::8000"))
	for _, p := range ps {
		fmt.Println(p)
	}
	// Output:
	// 2001:db8::1/128
	// 2001:db8::2/127
	// 2001:db8::4/126
	// 2001:db8::8/125
	// 2001:db8::10/124
	// 2001:db8::20/123
	// 2001:db8::40/122
	// 2001:db8::80/121
	// 2001:db8::100/120
	// 2001:db8::200/119
	// 2001:db8::400/118
	// 2001:db8::800/117
	// 2001:db8::1000/116
	// 2001:db8::2000/115
	// 2001:db8::4000/114
	// 2001:db8::8000/128
}

// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"net"
	"sort"
)

var (
	_ encoding.BinaryMarshaler   = &Prefix{}
	_ encoding.BinaryUnmarshaler = &Prefix{}
	_ encoding.TextMarshaler     = &Prefix{}
	_ encoding.TextUnmarshaler   = &Prefix{}

	_ fmt.Stringer = &Prefix{}
)

const (
	IPv4PrefixLen = 8 * net.IPv4len // maximum number of prefix length in bits
	IPv6PrefixLen = 8 * net.IPv6len // maximum number of prefix length in bits
)

// A Prefix represents an IP address prefix.
type Prefix struct {
	net.IPNet
}

func (p *Prefix) lastIPv4Int() ipv4Int {
	return ipToIPv4Int(p.IP) | ipv4Int(^mask32(p.Len()))
}

func (p *Prefix) lastIPv4MappedIPv6Int() ipv6Int {
	i4 := ipToIPv4Int(p.IP) | ipv4Int(^mask32(p.Len()))
	var i6 ipv6Int
	i6[1] = 0x0000ffff00000000 | uint64(i4)
	return i6
}

func (p *Prefix) lastIPv6Int() ipv6Int {
	i := ipToIPv6Int(p.IP)
	var m ipv6Int
	m.invmask(p.Len())
	i[0], i[1] = i[0]|m[0], i[1]|m[1]
	return i
}

// Equal reports whether the prefix p and q are equal.
func (p *Prefix) Equal(q *Prefix) bool {
	return p.IP.Equal(q.IP) && net.IP(p.Mask).Equal(net.IP(q.Mask))
}

// Exclude returns a list of prefixes that do not contain the prefix
// q.
func (p *Prefix) Exclude(q *Prefix) []Prefix {
	if !p.Contains(q.IP) {
		return nil
	}
	if p.Equal(q) {
		return []Prefix{*q}
	}
	subsFn := subnetsIPv6
	if p.IP.To4() != nil {
		subsFn = subnetsIPv4
	}
	var ps []Prefix
	l, r := subsFn(p, false)
	for !l.Equal(q) && !r.Equal(q) {
		if l.Contains(q.IP) {
			ps = append(ps, *r)
			l, r = subsFn(l, true)
		} else if r.Contains(q.IP) {
			ps = append(ps, *l)
			l, r = subsFn(r, true)
		}
	}
	if l.Equal(q) {
		ps = append(ps, *r)
	} else if r.Equal(q) {
		ps = append(ps, *l)
	}
	return ps
}

func subnetsIPv4(p *Prefix, reuse bool) (l *Prefix, r *Prefix) {
	i := ipToIPv4Int(p.IP) | ipv4Int(1<<uint(IPv4PrefixLen-p.Len()-1))
	r = i.prefix(p.Len()+1, IPv4PrefixLen)
	if reuse {
		l = p
		binary.BigEndian.PutUint32(l.Mask, mask32(l.Len()+1))
	} else {
		l = ipToIPv4Int(p.IP).prefix(p.Len()+1, IPv4PrefixLen)
	}
	return
}

func subnetsIPv6(p *Prefix, reuse bool) (l *Prefix, r *Prefix) {
	i := ipToIPv6Int(p.IP)
	id := ipv6Int{0, 1}
	id.lsh(IPv6PrefixLen - p.Len() - 1)
	ii := ipv6Int{i[0] | id[0], i[1] | id[1]}
	r = ii.prefix(p.Len()+1, IPv6PrefixLen)
	if reuse {
		l = p
		var m ipv6Int
		m.mask(l.Len() + 1)
		binary.BigEndian.PutUint64(l.Mask[:8], m[0])
		binary.BigEndian.PutUint64(l.Mask[8:16], m[1])
	} else {
		l = i.prefix(p.Len()+1, IPv6PrefixLen)
	}
	return
}

// Hostmask returns a host mask, the inverse mask of p's network mask.
func (p *Prefix) Hostmask() net.IPMask {
	return invert(p.Mask)
}

// Last returns the last IP in the address range of prefix p.
// It returns the address of the prefix when the prefix contains only
// one address.
func (p *Prefix) Last() net.IP {
	if p.IP.To4() != nil {
		i := p.lastIPv4Int()
		return i.ip()
	}
	if p.IP.To16() != nil && p.IP.To4() == nil {
		i := p.lastIPv6Int()
		return i.ip()
	}
	return nil
}

// Len returns the length of the prefix p in bits.
func (p *Prefix) Len() int {
	l, _ := p.Mask.Size()
	return l
}

// MarshalBinary returns a BGP NLRI binary form of the prefix p.
func (p *Prefix) MarshalBinary() ([]byte, error) {
	ip := p.IP
	if p.IP.To4() != nil {
		ip = p.IP.To4()
	}
	var b [1 + net.IPv6len]byte
	n := p.Len()
	l := n / 8
	if n%8 > 0 {
		l++
	}
	b[0] = byte(n)
	l++
	copy(b[1:l], ip)
	return b[:l], nil
}

// MarshalText returns a UTF-8-encoded text form of the prefix p.
func (p *Prefix) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// NumNodes returns the number of IP node addresses in the prefix p.
func (p *Prefix) NumNodes() *big.Int {
	i := new(big.Int).SetBytes(invert(p.Mask))
	return i.Add(i, big.NewInt(1))
}

// Overlaps reports whether the prefix p overlaps with the prefix q.
func (p *Prefix) Overlaps(q *Prefix) bool {
	return p.Contains(q.IP) || p.Contains(q.Last()) || q.Contains(p.IP) || q.Contains(p.Last())
}

func (p Prefix) String() string {
	return p.IPNet.String()
}

// Subnets returns a list of prefixes that are split from the prefix
// p, into small address blocks by n which represents a number of
// subnetworks in power of 2 notation.
func (p *Prefix) Subnets(n int) []Prefix {
	if 0 > n || n > 17 { // don't bother runtime.makeslice by big numbers
		return nil
	}
	ps := make([]Prefix, 1<<uint(n))
	if p.IP.To4() != nil {
		x := ipToIPv4Int(p.IP)
		off := uint(IPv4PrefixLen - p.Len() - n)
		for i := range ps {
			ii := x | ipv4Int(i<<off)
			ps[i] = *ii.prefix(p.Len()+n, IPv4PrefixLen)
		}
		return ps
	}
	x := ipToIPv6Int(p.IP)
	off := IPv6PrefixLen - p.Len() - n
	for i := range ps {
		id := ipv6Int{0, uint64(i)}
		id.lsh(off)
		ii := ipv6Int{x[0] | id[0], x[1] | id[1]}
		ps[i] = *ii.prefix(p.Len()+n, IPv6PrefixLen)
	}
	return ps
}

// UnmarshalBinary replaces the prefix p with the BGP NLRI binary form
// b.
func (p *Prefix) UnmarshalBinary(b []byte) error {
	if p.IP.To4() != nil {
		binary.BigEndian.PutUint32(p.Mask, mask32(int(b[0])))
		copy(p.IP, net.IPv4zero)
		copy(p.IP.To4(), b[1:])
	}
	if p.IP.To16() != nil && p.IP.To4() == nil {
		var m ipv6Int
		m.mask(int(b[0]))
		binary.BigEndian.PutUint64(p.Mask[:8], m[0])
		binary.BigEndian.PutUint64(p.Mask[8:16], m[1])
		copy(p.IP, net.IPv6unspecified)
		copy(p.IP, b[1:])
	}
	return nil
}

// UnmarshalText replaces the prefix p with txt.
func (p *Prefix) UnmarshalText(txt []byte) error {
	_, n, err := net.ParseCIDR(string(txt))
	if err != nil {
		return err
	}
	copy(p.IP.To16(), n.IP.To16())
	copy(p.Mask, n.Mask)
	return nil
}

// Aggregate aggregates the prefixes ps and returns a list of
// aggregated prefixes.
func Aggregate(ps []Prefix) []Prefix {
	ps = newSortedPrefixes(ps, sortDescending, true)
	switch len(ps) {
	case 0:
		return nil
	case 1:
		return ps[:1]
	}
	bfFn, superFn := branchingFactorIPv6, supernetIPv6
	if ps[0].IP.To4() != nil {
		bfFn, superFn = branchingFactorIPv4, supernetIPv4
	}
	var lastAggr *Prefix
	var djnts, aggrs []Prefix
	cands := make([]Prefix, 0, len(ps))
	for len(ps) > 0 {
		l := ps[0].Len()
		if l == 0 {
			djnts = append(djnts, ps[0])
			ps = ps[1:]
			continue
		}
		cands = cands[:0]
		for i := range ps {
			if ps[i].Len() != l {
				break
			}
			cands = append(cands, ps[i])
		}
		if lastAggr != nil && lastAggr.Len() == l {
			cands = append([]Prefix{*lastAggr}, cands...)
		}
		n, ok := bfFn(cands)
		if !ok {
			djnts = append(djnts, ps[0])
			ps = ps[1:]
			continue
		}
		aggr := superFn(cands[:n])
		if lastAggr != nil {
			ps = ps[n-1:]
			aggrs = aggrs[:len(aggrs)-1]
		} else {
			ps = ps[n:]
		}
		aggrs = append(aggrs, *aggr)
		lastAggr = aggr
	}
	aggrs = append(aggrs, djnts...)
	sort.Sort(byAscending(aggrs))
	return aggrs
}

func branchingFactorIPv4(ps []Prefix) (int, bool) {
	var lastBF, lastN int
	base := ipToIPv4Int(ps[0].IP.Mask(ps[0].Mask))
	mask := ipMaskToIPv4Int(ps[0].Mask)
	l := ps[0].Len()
	for bf := 1; bf < IPv4PrefixLen; bf++ {
		n, nfull := 0, 1<<uint(bf)
		max := ipv4Int(1 << uint(bf))
		aggrMask := mask << uint(bf)
		for pat := ipv4Int(0); pat < max; pat++ {
			aggr := base&aggrMask | pat<<uint(IPv4PrefixLen-l)
			for _, p := range ps {
				i := ipToIPv4Int(p.IP)
				if aggr == i&mask {
					n++
				}
			}
		}
		if n < nfull {
			break
		}
		lastBF = bf
		lastN = n
	}
	n := 1 << uint(lastBF)
	return n, lastN >= n
}

func branchingFactorIPv6(ps []Prefix) (int, bool) {
	var lastBF, lastN int
	base := ipToIPv6Int(ps[0].IP)
	mask := ipMaskToIPv6Int(ps[0].Mask)
	l := ps[0].Len()
	for bf := 1; bf < IPv6PrefixLen; bf++ {
		n, nfull := 0, 1<<uint(bf)
		pat, max := ipv6Int{0, 0}, ipv6Int{0, 1}
		max.lsh(bf)
		var aggrMask ipv6Int
		aggrMask.mask(l - bf)
		for ; pat.cmp(&max) < 0; pat.incr() {
			npat := pat
			npat.lsh(IPv6PrefixLen - l)
			var aggr ipv6Int
			aggr[0], aggr[1] = base[0]&aggrMask[0]|npat[0], base[1]&aggrMask[1]|npat[1]
			for _, p := range ps {
				i := ipToIPv6Int(p.IP)
				if aggr[0] == i[0]&mask[0] && aggr[1] == i[1]&mask[1] {
					n++
				}
			}
		}
		if n < nfull {
			break
		}
		lastBF = bf
		lastN = n
	}
	n := 1 << uint(lastBF)
	return n, lastN >= n
}

// Compare returns an integer comparing two prefixes.
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b.
func Compare(a, b *Prefix) int {
	return compareAscending(a, b)
}

// NewPrefix returns a new prefix.
func NewPrefix(n *net.IPNet) *Prefix {
	n.IP = n.IP.To16()
	return &Prefix{IPNet: *n}
}

// Summarize summarizes the address range from first to last and
// returns a list of prefixes.
func Summarize(first, last net.IP) []Prefix {
	if fip := first.To4(); fip != nil {
		lip := last.To4()
		if lip == nil {
			return nil
		}
		return summarizeIPv4(fip, lip)
	}
	if fip := first.To16(); fip != nil && fip.To4() == nil {
		lip := last.To16()
		if lip == nil || last.To4() != nil {
			return nil
		}
		return summarizeIPv6(fip, lip)
	}
	return nil
}

const ipv4IntEOR = ipv4Int(math.MaxUint32)

func summarizeIPv4(fip, lip net.IP) []Prefix {
	var ps []Prefix
	fi, li := ipToIPv4Int(fip), ipToIPv4Int(lip)
	for fi.cmp(li) <= 0 {
		n := IPv4PrefixLen
		for n > 0 {
			m := ipv4Int(mask32(n - 1))
			l, r := fi&m, fi|ipv4Int(^mask32(n-1))
			if fi.cmp(l) != 0 || r.cmp(li) > 0 {
				break
			}
			n--
		}
		p := fi.prefix(n, IPv4PrefixLen)
		ps = append(ps, *p)
		fi = p.lastIPv4Int()
		if fi == ipv4IntEOR {
			break
		}
		fi++
	}
	return ps
}

var ipv6IntEOR = ipv6Int{math.MaxUint64, math.MaxUint64}

func summarizeIPv6(fip, lip net.IP) []Prefix {
	var ps []Prefix
	fi, li := ipToIPv6Int(fip), ipToIPv6Int(lip)
	for fi.cmp(&li) <= 0 {
		n := IPv6PrefixLen
		for n > 0 {
			var m ipv6Int
			m.mask(n - 1)
			l, r := fi, fi
			l[0], l[1] = l[0]&m[0], l[1]&m[1]
			r.invmask(n - 1)
			r[0], r[1] = fi[0]|r[0], fi[1]|r[1]
			if fi.cmp(&l) != 0 || r.cmp(&li) > 0 {
				break
			}
			n--
		}
		p := fi.prefix(n, IPv6PrefixLen)
		ps = append(ps, *p)
		fi = p.lastIPv6Int()
		if fi[0] == ipv6IntEOR[0] && fi[1] == ipv6IntEOR[1] {
			break
		}
		fi.incr()
	}
	return ps
}

// Supernet finds out a shortest common prefix for the prefixes ps.
// It returns nil when no suitable prefix is found.
func Supernet(ps []Prefix) *Prefix {
	if len(ps) == 0 {
		return nil
	}
	if ps[0].IP.To4() != nil {
		ps = byAddrFamily(ps).newIPv4Prefixes()
	}
	if ps[0].IP.To16() != nil && ps[0].IP.To4() == nil {
		ps = byAddrFamily(ps).newIPv6Prefixes()
	}
	switch len(ps) {
	case 0:
		return nil
	case 1:
		return &ps[0]
	}
	if ps[0].IP.To4() != nil {
		return supernetIPv4(ps)
	}
	return supernetIPv6(ps)
}

func supernetIPv4(ps []Prefix) *Prefix {
	base := ipToIPv4Int(ps[0].IP.Mask(ps[0].Mask))
	mask := ipMaskToIPv4Int(ps[0].Mask)
	n := ps[0].Len()
	for _, p := range ps[1:] {
		i := ipToIPv4Int(p.IP)
		if diff := uint32((base ^ i) & mask); diff != 0 {
			if l := nlz32(diff); l < n {
				n = l
			}
		}
	}
	if n == 0 {
		return nil
	}
	return ipToPrefix(ps[0].IP, n, IPv4PrefixLen)
}

func supernetIPv6(ps []Prefix) *Prefix {
	base := ipToIPv6Int(ps[0].IP.Mask(ps[0].Mask))
	mask := ipMaskToIPv6Int(ps[0].Mask)
	n := ps[0].Len()
	var diff ipv6Int
	for _, p := range ps[1:] {
		i := ipToIPv6Int(p.IP)
		diff[0], diff[1] = (base[0]^i[0])&mask[0], (base[1]^i[1])&mask[1]
		if diff[0] != 0 {
			if l := nlz64(diff[0]); l < n {
				n = l
			}
		} else if diff[1] != 0 {
			if l := nlz64(diff[1]); 64+l < n {
				n = 64 + l
			}
		}
	}
	if n == 0 {
		return nil
	}
	return ipToPrefix(ps[0].IP, n, IPv6PrefixLen)
}

type ipv4Int uint32

func (i ipv4Int) cmp(j ipv4Int) int {
	if i < j {
		return -1
	}
	if i > j {
		return +1
	}
	return 0
}

func (i ipv4Int) ip() net.IP {
	ip := make(net.IP, net.IPv6len)
	copy(ip, net.IPv4zero)
	binary.BigEndian.PutUint32(ip.To4(), uint32(i))
	return ip.To16()
}

func (i ipv4Int) prefix(l, z int) *Prefix {
	ip := i.ip()
	m := net.CIDRMask(l, z)
	return &Prefix{IPNet: net.IPNet{IP: ip.Mask(m).To16(), Mask: m}}
}

type ipv6Int [2]uint64

func (i *ipv6Int) cmp(j *ipv6Int) int {
	if i[0] < j[0] {
		return -1
	}
	if i[0] > j[0] {
		return +1
	}
	if i[1] < j[1] {
		return -1
	}
	if i[1] > j[1] {
		return +1
	}
	return 0
}

func (i *ipv6Int) decr() {
	if i[0] == 0 && i[1] == 0 {
		return
	}
	if i[1] > 0 {
		i[1]--
	} else {
		i[0]--
		i[1] = math.MaxUint64
	}
}

func (i *ipv6Int) incr() {
	if i[1] == math.MaxUint64 {
		i[0]++
		i[1] = 0
	} else {
		i[1]++
	}
}

func (i *ipv6Int) invmask(n int) {
	if n > 64 {
		i[0], i[1] = ^mask64(64), ^mask64(n-64)
	} else {
		i[0], i[1] = ^mask64(n), mask64(64)
	}
}

func (i *ipv6Int) lsh(n int) {
	i[0] = i[0]<<uint(n) | i[1]>>uint(64-n) | i[1]<<uint(n-64)
	i[1] = i[1] << uint(n)
}

func (i *ipv6Int) mask(n int) {
	if n > 64 {
		i[0], i[1] = mask64(64), mask64(n-64)
	} else {
		i[0], i[1] = mask64(n), 0
	}
}

func (i *ipv6Int) ip() net.IP {
	ip := make(net.IP, net.IPv6len)
	binary.BigEndian.PutUint64(ip[:8], i[0])
	binary.BigEndian.PutUint64(ip[8:16], i[1])
	return ip
}

func (i *ipv6Int) prefix(l, z int) *Prefix {
	ip := i.ip()
	m := net.CIDRMask(l, z)
	return &Prefix{IPNet: net.IPNet{IP: ip.Mask(m), Mask: m}}
}

func invert(s []byte) []byte {
	d := make([]byte, len(s))
	for i := range s {
		d[i] = ^s[i]
	}
	return d
}

func ipToIPv4Int(ip net.IP) ipv4Int {
	return ipv4Int(binary.BigEndian.Uint32(ip.To4()))
}

func ipToIPv6Int(ip net.IP) ipv6Int {
	return ipv6Int{binary.BigEndian.Uint64(ip[:8]), binary.BigEndian.Uint64(ip[8:16])}
}

func ipMaskToIPv4Int(m net.IPMask) ipv4Int {
	return ipv4Int(binary.BigEndian.Uint32(m))
}

func ipMaskToIPv6Int(m net.IPMask) ipv6Int {
	return ipv6Int{binary.BigEndian.Uint64(m[:8]), binary.BigEndian.Uint64(m[8:16])}
}

func ipToPrefix(s net.IP, l, z int) *Prefix {
	d := make(net.IP, net.IPv6len)
	copy(d, s.To16())
	m := net.CIDRMask(l, z)
	return &Prefix{IPNet: net.IPNet{IP: d.Mask(m).To16(), Mask: m}}
}

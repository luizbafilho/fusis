package net

import (
	"net"
	"time"

	"github.com/mdlayher/arp"
)

var (
	ethernetBroadcast = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

func SendGratuitousARPReply(ip string, iface string) error {
	// Set up ARP client with socket
	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return err
	}

	c, err := arp.NewClient(ifi)
	if err != nil {
		return err
	}

	// Set request deadline from flag
	if err := c.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return err
	}

	srcIp := net.ParseIP(ip).To4()
	packet, err := arp.NewPacket(arp.OperationReply, ifi.HardwareAddr, srcIp, ethernetBroadcast, net.IPv4bcast)
	if err != nil {
		return err
	}

	if err := c.WriteTo(packet, ethernetBroadcast); err != nil {
		return err
	}

	// Clean up ARP client socket
	_ = c.Close()

	return nil
}

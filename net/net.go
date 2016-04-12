package net

import (
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func AddDefaultGateway(ip string) error {
	//TODO: Delete previous default gw
	err := netlink.RouteAdd(&netlink.Route{
		Scope: netlink.SCOPE_UNIVERSE,
		Gw:    net.ParseIP(ip),
	})
	if err != nil {
		log.Errorf("Adding Default Gateway: %s", ip)
		return err
	}
	return nil
}

func AddIp(ip string, iface string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(ip)
	if err != nil {
		return err
	}

	netlink.AddrAdd(link, addr)
	return nil
}

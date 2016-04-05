package main

import (
	"fmt"

	"net"

	"github.com/vishvananda/netlink"
)

func main() {
	// gwRoutes, err := netlink.RouteGet(net.ParseIP("10.0.0.1"))
	// if err != nil {
	// 	fmt.Println(err)
	// }
	err := netlink.RouteAdd(&netlink.Route{
		Scope: netlink.SCOPE_UNIVERSE,
		Gw:    net.ParseIP("192.168.33.1"),
	})
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println("===>", gwRoutes)
}

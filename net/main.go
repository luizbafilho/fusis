package main

import (
	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/engine"
)

func main() {
	// =======>> Adicionando default gateway
	// gwRoutes, err := netlink.RouteGet(net.ParseIP("10.0.0.1"))
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err := netlink.RouteAdd(&netlink.Route{
	// 	Scope: netlink.SCOPE_UNIVERSE,
	// 	Gw:    net.ParseIP("192.168.33.1"),
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("===>", gwRoutes)

	// cs := infra.NewCloudstackIaaS("0b5b922f-6b71-4955-b6bf-250685323dc9", "vr5P_5mC_H7vN1MDRQqotbW8h6EEjjnIGrDiqhLEyHJHY8lb_wznIDkeNPgjfmv45M4PCqkRX6fzxk5bMY_etQ", "rz7-Hek8YpblTb8wOXj-oaK6ZW2sAIF_Ph7Wy53q2GLLWNrAe1px3LAGW23OW3KanOUz1OHEatLOJb1WDK8Cvw")
	// cs.SetVip("fusis")
	client := api.NewClient("http://localhost:8000")

	// svc := engine.Service{
	// 	Host:      "10.2.3.9",
	// 	Port:      8081,
	// 	Name:      "blabla",
	// 	Protocol:  "tcp",
	// 	Scheduler: "lc",
	// }
	//
	// err := client.CreateService(svc)
	// if err != nil {
	// 	panic(err)
	// }

	dst := engine.Destination{
		Host:      "192.168.1.1",
		Port:      80,
		Weight:    1,
		Mode:      "nat",
		Name:      "hostname",
		ServiceId: "blabla",
	}

	err := client.AddDestination(dst)
	if err != nil {
		panic(err)
	}
}

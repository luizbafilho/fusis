package bgp

import (
	"log"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/state"
	bgp_config "github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/packet/bgp"
	gobgp "github.com/osrg/gobgp/server"
	"github.com/osrg/gobgp/table"
)

type BgpService struct {
	bgp    *gobgp.BgpServer
	config *config.BalancerConfig
}

type Syncer interface {
	Sync(state state.State) error
}

func NewBgpService(conf *config.BalancerConfig) (*BgpService, error) {
	err := validateConfig(conf)
	if err != nil {
		return nil, err
	}

	return &BgpService{
		bgp:    gobgp.NewBgpServer(),
		config: conf,
	}, nil
}

func validateConfig(conf *config.BalancerConfig) error {
	if _, err := govalidator.ValidateStruct(conf.Bgp); err != nil {
		return err
	}

	return nil
}

func (bs *BgpService) Serve() {
	go bs.bgp.Serve()

	// // global configuration
	// req := gobgp.NewGrpcRequest(gobgp.REQ_START_SERVER, "", bgp.RouteFamily(0), &gobgp_api.StartServerRequest{
	// 	Global: &gobgp_api.Global{
	// 		As:       bs.config.Bgp.As,
	// 		RouterId: bs.config.Bgp.RouterId,
	// 	},
	// })
	// bs.bgp.GrpcReqCh <- req
	// res := <-req.ResponseCh
	// if err := res.Err(); err != nil {
	// 	log.Fatal(err)
	// }
	// global configuration
	global := &bgp_config.Global{
		Config: bgp_config.GlobalConfig{
			As:       bs.config.Bgp.As,
			RouterId: bs.config.Bgp.RouterId,
		},
	}

	if err := bs.bgp.Start(global); err != nil {
		log.Fatal("Failed starting BGP service.", err)
	}

	for _, n := range bs.config.Bgp.Neighbors {
		bs.addNeighbor(n)
	}
}

func (bs *BgpService) addNeighbor(nb config.Neighbor) {
	// neighbor configuration
	// req := gobgp.NewGrpcRequest(gobgp.REQ_GRPC_ADD_NEIGHBOR, "", bgp.RouteFamily(0), &gobgp_api.AddNeighborRequest{
	// 	Peer: &gobgp_api.Peer{
	// 		Conf: &gobgp_api.PeerConf{
	// 			NeighborAddress: n.Address,
	// 			PeerAs:          n.PeerAs,
	// 		},
	// 		Transport: &gobgp_api.Transport{
	// 			LocalAddress: transportAddress,
	// 		},
	// 	},
	// })
	// bs.bgp.GrpcReqCh <- req
	// res := <-req.ResponseCh
	// if err := res.Err(); err != nil {
	// 	log.Fatal("Adding BGP Neighbor failed", err)
	// }
	// neighbor configuration
	n := &bgp_config.Neighbor{
		Config: bgp_config.NeighborConfig{
			NeighborAddress: nb.Address,
			PeerAs:          nb.PeerAs,
		},
	}

	if err := bs.bgp.AddNeighbor(n); err != nil {
		log.Fatal("Adding BGP Neighbor failed", err)
	}
}

func (bs *BgpService) AddPath(route string) {
	// add routes
	// path, _ := cmd.ParsePath(bgp.RF_IPv4_UC, []string{route})
	// req := gobgp.NewGrpcRequest(gobgp.REQ_ADD_PATH, "", bgp.RouteFamily(0), &gobgp_api.AddPathRequest{
	// 	Resource: gobgp_api.Resource_GLOBAL,
	// 	Path:     path,
	// })
	// bs.bgp.GrpcReqCh <- req
	// res := <-req.ResponseCh
	// if err := res.Err(); err != nil {
	// 	log.Fatal(err)
	// }
	attrs := []bgp.PathAttributeInterface{
		bgp.NewPathAttributeOrigin(0),
		bgp.NewPathAttributeNextHop("0.0.0.0"),
	}
	if _, err := bs.bgp.AddPath("", []*table.Path{table.NewPath(nil, bgp.NewIPAddrPrefix(32, route), false, attrs, time.Now(), false)}); err != nil {
		log.Fatal(err)
	}
}

func (bs *BgpService) delPath(route string) {
	// del routes
	// path, _ := cmd.ParsePath(bgp.RF_IPv4_UC, []string{route})
	// req := gobgp.NewGrpcRequest(gobgp.REQ_DELETE_PATH, "", bgp.RouteFamily(0), &gobgp_api.DeletePathRequest{
	// 	Resource: gobgp_api.Resource_GLOBAL,
	// 	Path:     path,
	// })
	// bs.bgp.GrpcReqCh <- req
	// res := <-req.ResponseCh
	// if err := res.Err(); err != nil {
	// 	log.Fatal(err)
	// }
}

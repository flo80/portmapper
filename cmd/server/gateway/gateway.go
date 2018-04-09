package gateway

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jackpal/gateway"
	natpmp "github.com/jackpal/go-nat-pmp"
	"github.com/metricube/upnp"
)

var (
	defaultLifetime = 600
	defaultProtocol = "tcp"
)

type Gateway struct {
	natPMPfails   bool
	upnpfails     bool
	natpmpGateway net.IP
	natpmpClient  natpmp.Client
	upnpGateway   upnp.UPNP
}

func NewGateway() (*Gateway, error) {
	g := Gateway{}
	g.findPMPGateway()
	g.findUPNPgateway()

	if g.natPMPfails && g.upnpfails {
		return nil, fmt.Errorf("no gateways can be found")
	}

	if !g.natPMPfails {
		g.natpmpClient = *natpmp.NewClient(g.natpmpGateway)
	}

	return &g, nil
}

func (g *Gateway) findPMPGateway() {
	r, err := gateway.DiscoverGateway()
	if err != nil {
		g.natPMPfails = true
		return
	}
	g.natpmpGateway = r
}

func (g *Gateway) findUPNPgateway() {
	r, err := upnp.NewUPNP()
	if err != nil {
		g.upnpfails = true
		return
	}
	g.upnpGateway = *r
}

/*
func (g *Gateway) getNatpmpGateway() *net.IP {
	if g.natPMPfails {
		return nil
	}
	if g.natpmpGateway == nil {
		g.findPMPGateway()
	}
	return &g.natpmpGateway
}
func (g *Gateway) getUPNPgateway() *upnp.UPNP {
	if g.upnpGateway.Gateway == nil {
		g.findUPNPgateway()
	}
	return &g.upnpGateway
}
*/

func (g *Gateway) OpenPortNatPMP(internalPort, externalPort int, protocol string, lifetime int) error {
	if lifetime < 1 {
		return fmt.Errorf("Lifetime needs to be >= 1")
	}
	return g.portMapNatPMP(internalPort, externalPort, protocol, lifetime)
}

func (g *Gateway) ClosePortNatPMP(internalPort, externalPort int) error {
	return g.portMapNatPMP(internalPort, externalPort, defaultProtocol, 0)
}

func (g *Gateway) portMapNatPMP(internalPort, externalPort int, protocol string, lifetime int) error {
	if g.natPMPfails {
		return fmt.Errorf("NAT PMP not available")
	}

	res, err := g.natpmpClient.AddPortMapping(protocol, internalPort, externalPort, lifetime)
	if err != nil {
		log.Printf("NAT PMP could not map port %v", err)
		return err
	}

	log.Printf(`Port mapping for NAT-PMP
	External Port: %v
	Internal Port: %v
	Mapping Lifetime: %vs
	`,
		res.MappedExternalPort, res.InternalPort, res.PortMappingLifetimeInSeconds)

	return nil
}

func (g *Gateway) OpenPortUPNP(internalPort, externalPort int, protocol string, lifetime int) error {
	err := g.upnpGateway.AddPortMapping(internalPort, externalPort, protocol)
	if err != nil {
		log.Printf("UPNP could not map port %v", err)
		return err
	}
	log.Printf("Port mapping for UPNP successful\n")

	go func() {
		duration, err := time.ParseDuration(fmt.Sprintf("%vs", lifetime))
		if err != nil {
			log.Printf("parsing of duration failed %v", err)
		}
		time.Sleep(duration)
		_ = g.ClosePortMapUPNP(externalPort)
	}()

	return nil
}

func (g *Gateway) ClosePortMapUPNP(externalPort int) error {
	err := g.upnpGateway.DelPortMapping(externalPort, "tcp")
	if err != nil {
		log.Printf("could not delete port map %v", err)
		return err
	}
	log.Printf("port mapping UPNP deleted\n")
	return nil
}

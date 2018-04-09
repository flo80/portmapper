package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/flo80/portmapping/api"
	"github.com/flo80/portmapping/cmd/server/gateway"
)

const defaultGRPCIP = ""
const defaultGRPCPort = 7777

type Server struct {
	gateway *gateway.Gateway
}

func main() {
	addressServer := flag.String("server", fmt.Sprintf("%v:%v", defaultGRPCIP, defaultGRPCPort), "Set GRPC server ip:port")
	flag.Parse()

	server := Server{}
	g, err := gateway.NewGateway()
	if err != nil {
		log.Fatalf("could not get a gateway: %v", err)
	}
	server.gateway = g

	lis, err := net.Listen("tcp", *addressServer)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	api.RegisterNATserviceServer(grpcServer, &server)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func (s *Server) OpenPort(ctx context.Context, req *api.OpenPortRequest) (*api.StatusResponse, error) {
	log.Printf("Open Port called: %v", req)
	internalPort := int(req.LocalPort)
	externalPort := int(req.ExternalPort)
	protocol := req.GetProtocol().String()
	lifetime := int(req.Lifetime)

	resp := &api.StatusResponse{
		Success:    api.Success_ok,
		Message:    "opening of port successful",
		Technology: req.Technology,
	}

	switch req.Technology {
	case api.Technology_natpmp:
		err := s.gateway.OpenPortNatPMP(internalPort, externalPort, protocol, lifetime)
		if err != nil {
			resp.Success = api.Success_notOk
			resp.Message = err.Error()
		}
		resp.Technology = api.Technology_natpmp

	case api.Technology_upnp:
		err := s.gateway.OpenPortUPNP(internalPort, externalPort, protocol, lifetime)
		if err != nil {
			resp.Success = api.Success_notOk
			resp.Message = err.Error()
		}
		resp.Technology = api.Technology_upnp
	default:
		log.Printf("tech not yet implemented")
		return nil, fmt.Errorf("technology not yet implemented")
	}
	return resp, nil
}

func (s *Server) ClosePort(ctx context.Context, req *api.ClosePortRequest) (*api.StatusResponse, error) {
	log.Printf("Close Port called: %v", req)
	internalPort := int(req.LocalPort)
	externalPort := int(req.ExternalPort)

	resp := &api.StatusResponse{
		Success:    api.Success_ok,
		Message:    "closing of port successful",
		Technology: req.Technology,
	}

	switch req.Technology {
	case api.Technology_natpmp:
		err := s.gateway.ClosePortNatPMP(internalPort, externalPort)
		if err != nil {
			resp.Success = api.Success_notOk
			resp.Message = err.Error()
		}
		resp.Technology = api.Technology_natpmp

	case api.Technology_upnp:
		err := s.gateway.ClosePortMapUPNP(externalPort)
		if err != nil {
			resp.Success = api.Success_notOk
			resp.Message = err.Error()
		}
		resp.Technology = api.Technology_upnp
	default:
		log.Printf("tech not yet implemented")
		return nil, fmt.Errorf("technology not yet implemented")
	}
	return resp, nil
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/flo80/portmapper/api"
	"github.com/flo80/portmapper/cmd/client/tools"
)

const defaultGRPCIP = "127.0.0.1"
const defaultGRPCPort = 7777

func main() {
	flag.Usage = tools.Usage
	addressServer := flag.String("server", fmt.Sprintf("%v:%v", defaultGRPCIP, defaultGRPCPort), "Set GRPC server ip:port")
	flag.Parse()

	fct, internalPort, externalPort, protocol, lifetime := tools.ParseArgs(flag.Args())
	fmt.Printf("Using GRPC server %v \n", *addressServer)

	conn, err := grpc.Dial(*addressServer, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := api.NewNATserviceClient(conn)

	tech := api.Technology_all
	proto := api.Protocol_tcp
	if protocol == "udp" {
		proto = api.Protocol_udp
	}

	switch fct {
	case "openpmp", "closepmp":
		tech = api.Technology_natpmp
	case "openupnp", "closeupnp":
		tech = api.Technology_upnp
	}

	switch fct {
	case "open", "openpmp", "openupnp":
		req := &api.OpenPortRequest{
			LocalPort:    int32(internalPort),
			ExternalPort: int32(externalPort),
			Protocol:     proto,
			Lifetime:     int32(lifetime),
			Technology:   tech,
		}

		res, err := c.OpenPort(context.Background(), req)
		if err != nil {
			log.Fatalf("opening failed: %v", err)
		}

		if res.Success == api.Success_ok {
			fmt.Printf("Success: %v", res)
			os.Exit(0)
		}
		fmt.Printf("Not successful: %v", res.Message)
		os.Exit(1)

	case "close", "closepmp", "closeupnp":
		req := &api.ClosePortRequest{
			LocalPort:    int32(internalPort),
			ExternalPort: int32(externalPort),
			Technology:   tech,
		}

		res, err := c.ClosePort(context.Background(), req)
		if err != nil {
			log.Fatalf("closing failed: %v", err)
		}

		if res.Success == api.Success_ok {
			fmt.Printf("Success: %v", res)
			os.Exit(0)
		}
		fmt.Printf("Not successful: %v", res.Message)
		os.Exit(1)

	default:
		fmt.Printf("function %v not implemented \n\n", fct)
		tools.Usage()
	}

}

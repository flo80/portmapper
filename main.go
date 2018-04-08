package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gateway "github.com/jackpal/gateway"
	natpmp "github.com/jackpal/go-nat-pmp"
	upnp "github.com/metricube/upnp"
)

const defaultLifetime = 600
const defaultProtocol = "tcp"

func usage() {
	fmt.Printf(`Usage:
%v command internalPort externalPort [protocol] [lifetime]

Commands:  
- open for opening up a port - tries NAT PMP and UPNP 
- close for closing a port - tries NAT PMP and UPNP
- openPMP & closePMP for only using NAT PMP
- openUPNP & closeUPNP for only using UPNP

Arguments:mkdir 
- Internal and external port in the range of 0-65535
- Protocol can be tcp or udp (default %v)
- Lifetime in secondes of port mapping 
	- for PMP guaranteed in protocol, for UPNP only guaranteed if app is not killed 
	- default lifetime for PMP is %v seconds, for UPNP unlimited (i.e. needs to be closed)

Exit Codes:
0 if mapping successful
1 if mapping no successful 
2 if arguments wrong
`, os.Args[0], defaultProtocol, defaultLifetime)
	os.Exit(2)
}

func main() {
	log.Printf("len: %v", len(os.Args))
	if len(os.Args) < 4 || len(os.Args) > 5 {
		usage()
	}

	fct := strings.ToLower(os.Args[1])
	internalPort, err := strconv.Atoi(os.Args[2])
	if err != nil || internalPort < 0 || internalPort > 65535 {
		usage()
	}

	externalPort, err := strconv.Atoi(os.Args[3])
	if err != nil || externalPort < 0 || externalPort > 65535 {
		usage()
	}

	protocol := defaultProtocol
	lifetime := defaultLifetime
	lifetimeSet := false
	if len(os.Args) > 4 {
		if os.Args[4] == "tcp" || os.Args[4] == "udp" {
			protocol = os.Args[4]
		} else {
			lifetime, err = strconv.Atoi(os.Args[4])
			lifetimeSet = true
			if err != nil || lifetime < 0 {
				usage()
			}
		}
	}
	if len(os.Args) == 6 {
		lifetime, err = strconv.Atoi(os.Args[5])
		lifetimeSet = true
		if err != nil || lifetime < 0 {
			usage()
		}

	}

	log.Printf(`Using Values:
	Command:       %v
	Internal Port: %v
	External Port: %v
	Protocol:      %v
	Lifetime:      %v seconds
	`, fct, internalPort, externalPort, protocol, lifetime)

	switch fct {
	case "open":
		err := portMapNatPMP(internalPort, externalPort, protocol, lifetime)
		if err != nil {
			err2 := portMapUPNP(internalPort, externalPort, protocol)
			if err2 != nil {
				fmt.Printf("Execution failed: %v, %v", err, err2)
				os.Exit(1)
			}
			if lifetimeSet {
				duration, err := time.ParseDuration(fmt.Sprintf("%vs", lifetime))
				if err != nil {
					fmt.Printf("Closing of port failed: %v", err)
					os.Exit(0)
				}
				time.Sleep(duration)
				err = deletePortMapUPNP(externalPort)
				if err != nil {
					fmt.Printf("Closing of port failed: %v", err)
					os.Exit(0)
				}
			}
		}
	case "close":
		err := portMapNatPMP(internalPort, externalPort, protocol, 0)
		if err != nil {
			err2 := deletePortMapUPNP(externalPort)
			if err2 != nil {
				fmt.Printf("Execution failed: %v, %v", err, err2)
				os.Exit(1)
			}
		}
	case "openpmp":
		err := portMapNatPMP(internalPort, externalPort, protocol, lifetime)
		if err != nil {
			fmt.Printf("Execution failed: %v", err)
			os.Exit(1)
		}
	case "closepmp":
		err := portMapNatPMP(internalPort, externalPort, protocol, 0)
		if err != nil {
			fmt.Printf("Execution failed: %v", err)
			os.Exit(1)
		}
	case "openupnp":
		err := portMapUPNP(internalPort, externalPort, protocol)
		if err != nil {
			fmt.Printf("Execution failed: %v", err)
			os.Exit(1)
		}
		if lifetimeSet {
			duration, err := time.ParseDuration(fmt.Sprintf("%vs", lifetime))
			if err != nil {
				fmt.Printf("Closing of port failed: %v", err)
				os.Exit(0)
			}
			time.Sleep(duration)
			err = deletePortMapUPNP(externalPort)
			if err != nil {
				fmt.Printf("Closing of port failed: %v", err)
				os.Exit(0)
			}
		}
	case "closepnp":
		err := deletePortMapUPNP(externalPort)
		if err != nil {
			fmt.Printf("Execution failed: %v", err)
			os.Exit(1)
		}
	default:
		usage()
	}
}

func portMapNatPMP(internalPort, externalPort int, protocol string, lifetime int) error {
	gatewayIP, err := gateway.DiscoverGateway()
	if err != nil {
		log.Printf("NAT PMP could not find gateway %v", err)
		return err
	}
	client := natpmp.NewClient(gatewayIP)
	response, err := client.GetExternalAddress()
	if err != nil {
		log.Printf("NAT PMP could not get external ip %v", err)
		return err
	}

	res, err := client.AddPortMapping("tcp", internalPort, externalPort, lifetime)
	if err != nil {
		log.Printf("NAT PMP could not map port %v", err)
		return err
	}

	log.Printf(`Port mapping for NAT-PMP
	External IP Address: %v.%v.%v.%v
	External Port: %v
	Internal Port: %v
	Mapping Lifetime: %vs
	`,
		response.ExternalIPAddress[0], response.ExternalIPAddress[1], response.ExternalIPAddress[2], response.ExternalIPAddress[3],
		res.MappedExternalPort, res.InternalPort, res.PortMappingLifetimeInSeconds)

	return nil
}

func portMapUPNP(internalPort, externalPort int, protocol string) error {
	client, err := upnp.NewUPNP()
	if err != nil {
		log.Printf("UPNP could not find gateway %v", err)
		return err
	}
	ip, err := client.ExternalIPAddress()
	if err != nil {
		log.Printf("UPNP could not get external ip %v", err)
		return err
	}
	err = client.AddPortMapping(internalPort, externalPort, "tcp")
	if err != nil {
		log.Printf("UPNP could not map port %v", err)
		return err
	}
	log.Printf(`Port mapping for UPNP
		External IP Address: %v
		`, ip)

	return nil
}

func deletePortMapUPNP(externalPort int) error {
	client, err := upnp.NewUPNP()
	if err != nil {
		log.Printf("could not find gateway %v", err)
		return err
	}

	client.DelPortMapping(externalPort, "tcp")
	if err != nil {
		log.Printf("could not delete port map %v", err)
		return err
	}
	log.Printf("port mapping deleted\n")
	return nil
}

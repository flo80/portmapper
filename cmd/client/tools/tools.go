package tools

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const defaultLifetime = 600
const defaultProtocol = "tcp"

func Usage() {
	fmt.Printf(`Usage:
%v command internalPort externalPort [protocol] [lifetime]

Commands:  
	- open for opening up a port - tries NAT PMP and UPNP 
	- close for closing a port - tries NAT PMP and UPNP
	- openPMP & closePMP for only using NAT PMP
	- openUPNP & closeUPNP for only using UPNP

Arguments: 
	- Internal and external port in the range of 0-65535
	- Protocol can be tcp or udp (default %v)
	- Lifetime in secondes of port mapping 
		- for PMP guaranteed in protocol, for UPNP only guaranteed if app is not killed 
		- default lifetime for PMP is %v seconds, for UPNP unlimited (i.e. needs to be closed)

Flags:
	`, os.Args[0], defaultProtocol, defaultLifetime)

	flag.PrintDefaults()

	fmt.Printf(`
Exit Codes:
	0 if mapping successful
	1 if mapping no successful 
	2 if arguments wrong
`)
	os.Exit(2)
}

func ParseArgs(args []string) (fct string, internalPort, externalPort int, protocol string, lifetime int) {
	log.Printf("len: %v", len(args))
	if len(args) < 3 || len(args) > 5 {
		Usage()
	}

	fct = strings.ToLower(args[0])
	internalPort, err := strconv.Atoi(args[1])
	if err != nil || internalPort < 0 || internalPort > 65535 {
		Usage()
	}

	externalPort, err = strconv.Atoi(args[2])
	if err != nil || externalPort < 0 || externalPort > 65535 {
		Usage()
	}

	protocol = defaultProtocol
	lifetime = defaultLifetime
	if len(args) > 3 {
		if args[3] == "tcp" || args[3] == "udp" {
			protocol = args[3]
		} else {
			lifetime, err = strconv.Atoi(args[3])
			if err != nil || lifetime < 0 {
				Usage()
			}
		}
	}
	if len(args) == 5 {
		lifetime, err = strconv.Atoi(args[4])
		if err != nil || lifetime < 0 {
			Usage()
		}

	}

	log.Printf(`Using Values:
	Command:       %v
	Internal Port: %v
	External Port: %v
	Protocol:      %v
	Lifetime:      %v seconds
	`, fct, internalPort, externalPort, protocol, lifetime)

	return fct, internalPort, externalPort, protocol, lifetime
}

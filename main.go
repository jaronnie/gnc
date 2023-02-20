package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var (
	ProtocolPorts []string

	LifeCycleTime int // second

	IsListen bool
)

type ProtocolPort struct {
	Protocol string
	Port     string
}

type Listener struct {
	Listener net.Listener

	Protocol string
}

func main() {
	bindFlags()

	checkPorts := preprocessing(ProtocolPorts)

	if LifeCycleTime != 0 && IsListen {
		time.AfterFunc(time.Duration(LifeCycleTime)*time.Second, func() {
			os.Exit(0)
		})
	}

	var conflictPorts []string

	var listeners []Listener

	for _, v := range checkPorts {
		var listen net.Listener
		var err error
		switch v.Protocol {
		case "tcp":
			listen, err = net.Listen("tcp", fmt.Sprintf(":%s", v.Port))
		case "udp":
			// todo
		}
		if err != nil {
			conflictPorts = append(conflictPorts, v.Port)
		} else {
			listeners = append(listeners, Listener{
				Listener: listen,
				Protocol: v.Protocol,
			})
		}
	}

	if len(conflictPorts) != 0 {
		for _, v := range listeners {
			_ = v.Listener.Close()
		}
		fmt.Printf("conflict ports: %s", strings.Join(conflictPorts, ","))
		return
	} else if !IsListen {
		return
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("success"))
	})

	for _, v := range listeners {
		v := v
		go func() {
			switch v.Protocol {
			case "tcp":
				_ = http.Serve(v.Listener, nil)
			}
		}()
	}

	select {}
}

func preprocessing(protoPorts []string) []ProtocolPort {
	var protocolPorts []ProtocolPort
	for _, v := range protoPorts {
		var protocol, port string
		split := strings.Split(v, "/")
		if len(split) == 2 {
			port = split[0]
			protocol = strings.ToLower(split[1])
		} else {
			port = split[0]
			protocol = "tcp"
		}
		protocolPorts = append(protocolPorts, ProtocolPort{
			Protocol: protocol,
			Port:     port,
		})
	}
	return protocolPorts
}

func bindFlags() {
	pflag.StringSliceVarP(&ProtocolPorts, "ports", "p", nil, "set ports to pre check, default protocol is tcp")
	pflag.IntVarP(&LifeCycleTime, "lifecycle-time", "t", 0, "set listen lifecycle time when listen")
	pflag.BoolVarP(&IsListen, "listen", "l", false, "is listen")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
}

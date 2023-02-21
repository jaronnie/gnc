package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kevwan/mapreduce"
	"github.com/spf13/pflag"
)

var (
	ProtocolPorts []string
	LifeCycleTime int // second
	IsListen      bool
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

	sig := make(chan os.Signal, 1)
	if LifeCycleTime != 0 && IsListen {
		time.AfterFunc(time.Duration(LifeCycleTime)*time.Second, func() {
			sig <- syscall.SIGQUIT
		})
	}

	var conflictPorts []string
	var listeners []Listener

	_ = mapreduce.MapReduceVoid(func(source chan<- interface{}) {
		for _, v := range checkPorts {
			source <- v
		}
	}, func(item interface{}, writer mapreduce.Writer, cancel func(error)) {
		v := item.(ProtocolPort)

		var listen net.Listener
		var err error
		switch v.Protocol {
		case "tcp":
			listen, err = net.Listen("tcp", fmt.Sprintf(":%s", v.Port))
		case "udp":
			// todo
		}
		if err != nil {
			writer.Write(v.Port)
		} else {
			writer.Write(Listener{
				Listener: listen,
				Protocol: v.Protocol,
			})
		}
	}, func(pipe <-chan interface{}, cancel func(error)) {
		for v := range pipe {
			if value, ok := v.(string); ok && value != "" {
				conflictPorts = append(conflictPorts, value)
			} else {
				if value, ok := v.(Listener); ok {
					listeners = append(listeners, value)
				}
			}
		}
	})

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
			case "udp":
				// todo
			}
		}()
	}

	signalHandler(listeners, sig)
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

func signalHandler(listeners []Listener, sig chan os.Signal) {
	// signal handler
	signal.Notify(sig, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-sig
		fmt.Println(fmt.Sprintf("get a signal %s", s.String()))
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			for _, v := range listeners {
				_ = v.Listener.Close()
			}
			fmt.Printf("close network successfully")
			return
		default:
			return
		}
	}
}

func bindFlags() {
	pflag.StringSliceVarP(&ProtocolPorts, "ports", "p", nil, "set ports to pre check, default protocol is tcp")
	pflag.IntVarP(&LifeCycleTime, "lifecycle-time", "t", 0, "set listen lifecycle time when listen")
	pflag.BoolVarP(&IsListen, "listen", "l", false, "is listen")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
}

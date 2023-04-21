package main

import (
	"flag"
	"log"

	"github.com/OmarTariq612/go-wstunnel/client"
	"github.com/OmarTariq612/go-wstunnel/server"
)

func main() {
	serverAddrOption := flag.String("s", "", "run as server, listen on [localip:]localport")
	tunnelAddrOption := flag.String("t", "", "run as tunnel client, specify [localip:]localport:host:port")

	flag.Parse()

	// server
	if *serverAddrOption != "" {
		srv := server.NewServer(*serverAddrOption)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	} else { // client
		cli := client.NewClient(*tunnelAddrOption, flag.Arg(0))
		if err := cli.Start(); err != nil {
			log.Println(err)
		}
	}
}

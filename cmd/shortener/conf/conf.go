package conf

import (
	"flag"
)

type config struct {
	serverAddress      string
	responseURLAddress string
}

var cnf config

func ParseFlags() {
	flag.StringVar(&cnf.serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cnf.responseURLAddress, "b", "http://localhost:8080", "address and port to response originalURL")

	flag.Parse()
}

func GetServAdr() string {
	return cnf.serverAddress
}

func GetResURLAdr() string {
	return cnf.responseURLAddress
}

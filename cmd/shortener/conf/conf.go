package conf

import (
	"flag"
)

type config struct {
	serverAddress        string
	responseUrlAddress   string
	schemeForResponseUrl string
}

var cnf config

func ParseFlags() {
	cnf.schemeForResponseUrl = "http"
	flag.StringVar(&cnf.serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cnf.responseUrlAddress, "b", "localhost:8080", "address and port to response URL")

	flag.Parse()
}

func GetServAdr() string {
	return cnf.serverAddress
}

func GetResUrlAdr() string {
	return cnf.responseUrlAddress
}

func GetSchemeResUrl() string {
	return cnf.schemeForResponseUrl
}

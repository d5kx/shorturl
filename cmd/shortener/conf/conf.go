package conf

import (
	"flag"
)

type config struct {
	serverAddress        string
	responseURLAddress   string
	schemeForResponseURL string
}

var cnf config

func ParseFlags() {
	cnf.schemeForResponseURL = "http"
	flag.StringVar(&cnf.serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cnf.responseURLAddress, "b", "localhost:8080", "address and port to response URL")

	flag.Parse()
}

func GetServAdr() string {
	return cnf.serverAddress
}

func GetResURLAdr() string {
	return cnf.responseURLAddress
}

func GetSchemeResURL() string {
	return cnf.schemeForResponseURL
}

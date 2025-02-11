package conf

import (
	"flag"
	"log"
	"net/url"
	"os"
)

type flags struct {
	flagServerAddress      string
	flagResponseURLAddress string
}

var cnf flags

func ParseFlags() {
	flag.StringVar(&cnf.flagServerAddress, "a", "localhost:8080", "address and port to start the HTTP server")
	flag.StringVar(&cnf.flagResponseURLAddress, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	flag.Parse()

	_, err := url.Parse(cnf.flagResponseURLAddress)
	if err != nil {
		log.Println("can't parse the base address of the resulting shortened URL (" + cnf.flagResponseURLAddress + "), set http://localhost:8080")
		cnf.flagResponseURLAddress = "http://localhost:8080"
	}

	stringVarEnv(&cnf.flagServerAddress, "SERVER_ADDRESS")
	stringVarEnv(&cnf.flagResponseURLAddress, "BASE_URL")
}

func GetServAdr() string {
	return cnf.flagServerAddress
}

func GetResURLAdr() string {
	return cnf.flagResponseURLAddress
}

func stringVarEnv(p *string, name string) {
	if v := os.Getenv(name); v != "" {
		*p = v
	}
}

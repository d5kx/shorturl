package conf

import (
	"flag"
	"net/url"
	"os"
)

type flags struct {
	flagServerAddress      string
	flagResponseURLAddress string
	flagLoggerLevel        string
}

var cnf flags

func ParseFlags() {
	flag.StringVar(&cnf.flagServerAddress, "a", "localhost:8080", "address and port to start the HTTP servers")
	flag.StringVar(&cnf.flagResponseURLAddress, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	flag.StringVar(&cnf.flagLoggerLevel, "l", "info", "loggers level")

	flag.Parse()

	stringVarEnv(&cnf.flagServerAddress, "SERVER_ADDRESS")
	stringVarEnv(&cnf.flagResponseURLAddress, "BASE_URL")

	_, err := url.Parse(cnf.flagResponseURLAddress)
	if err != nil {
		//loggers.Println("can't parse the base address of the resulting shortened URL (" + cnf.flagResponseURLAddress + "), set http://localhost:8080")
		cnf.flagResponseURLAddress = "http://localhost:8080"
	}

	if cnf.flagLoggerLevel != "info" && cnf.flagLoggerLevel != "debug" {
		cnf.flagLoggerLevel = "info"
	}
}

func GetServAdr() string {
	return cnf.flagServerAddress
}

func GetResURLAdr() string {
	return cnf.flagResponseURLAddress
}

func GetLoggerLevel() string {
	return cnf.flagLoggerLevel
}

func stringVarEnv(p *string, name string) {
	if v := os.Getenv(name); v != "" {
		*p = v
	}
}

package conf

import (
	"flag"
	"net/url"
	"os"
)

type flags struct {
	flagServerAddress              string
	flagResponseURLAddress         string
	flagLoggerLevel                string
	flagDBFileName                 string
	flagPostgreSQLConnectionString string
}

var cnf flags

func ParseFlags() {
	flag.StringVar(&cnf.flagServerAddress, "a", "localhost:8080", "address and port to start the HTTP servers")
	flag.StringVar(&cnf.flagResponseURLAddress, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	flag.StringVar(&cnf.flagLoggerLevel, "l", "debug", "loggers level")
	flag.StringVar(&cnf.flagDBFileName, "f", "tmp/short-url-db.json", "full file name to save DB")
	flag.StringVar(&cnf.flagPostgreSQLConnectionString, "d", "host=localhost port=5432 user=postgres password=820610 dbname=shorturl sslmode=disable", "connection string for PostgreSQL DB")

	flag.Parse()

	stringVarEnv(&cnf.flagServerAddress, "SERVER_ADDRESS")
	stringVarEnv(&cnf.flagResponseURLAddress, "BASE_URL")
	stringVarEnv(&cnf.flagDBFileName, "FILE_STORAGE_PATH")
	stringVarEnv(&cnf.flagPostgreSQLConnectionString, "DATABASE_DSN")

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

func GetDBFileName() string {
	return cnf.flagDBFileName
}

func GetPostgreSQLConnectionString() string { return cnf.flagPostgreSQLConnectionString }

func stringVarEnv(p *string, name string) {
	if v := os.Getenv(name); v != "" {
		*p = v
	}
}

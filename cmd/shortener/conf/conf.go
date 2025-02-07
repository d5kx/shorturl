package conf

import (
	"flag"
)

type config struct {
	serverAddress      string
	responseUrlAddress string
}

var Config config

func (c *config) ParseFlags() {

	flag.StringVar(&c.serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.responseUrlAddress, "b", "localhost:8080", "address and port to response URL")

	flag.Parse()
}

func (c *config) GetServAdr() string {
	return c.serverAddress
}

func (c *config) GetResUrlAdr() string {
	return c.responseUrlAddress
}

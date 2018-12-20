package common

import (
	"github.com/smartswarm/go/log"
	"strconv"
	"strings"
)

type HostAddress struct {

	FullAddress string
	IpAddress string
	Port int
}

func ComposeHostAddress(address string) *HostAddress {

	log.I2("ComposeHostAddress by string: %s", address)
	colonIndex := strings.Index(address, ":")
	if colonIndex < 0 {

		log.W("Address doesn't contain colon.")
		return nil
	}

	portString := address[colonIndex+1:]
	result := new(HostAddress)
	result.FullAddress = address
	result.IpAddress = address[:colonIndex]
	result.Port, _ = strconv.Atoi(portString)

	log.I2("IpAddress=[%s], Port=[%d]", result.IpAddress, result.Port)

	return result
}


func (this *HostAddress) IsValid() bool {

	return len(this.IpAddress) > 0 && this.Port > 0
}

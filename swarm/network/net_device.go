package network

import "fmt"

type NetDevice struct {

	HostUrl string		// The full URL, e.g. 127.0.0.1:8888
	IPAddress string
	Port int32
	Signature string

	PublicKey string
	PrivateKey string
	TempPublicKey string

}

func CreateDevice(ipAddress string, port int32) *NetDevice {

	device := new(NetDevice)
	device.IPAddress = ipAddress
	device.Port = port
	device.HostUrl = fmt.Sprintf("%s:%d", ipAddress, port)

	return device
}
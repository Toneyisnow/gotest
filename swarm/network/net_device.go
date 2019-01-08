package network

import "fmt"

type NetDevice struct {

	//// HostUrl string	   	// The full URL, e.g. 127.0.0.1:8888

	Id string 			 `json:"id"`
	IPAddress string     `json:"ip_address"`
	Port int32           `json:"port"`

	PublicKey []byte     `json:"public_key"`
	PrivateKey []byte	 `json:"private_key"`       // For a given config, only the Self node contains the Private key
	TempPublicKey []byte `json:"temp_public_key"`

}

func CreateDevice(ipAddress string, port int32) *NetDevice {

	device := new(NetDevice)
	device.IPAddress = ipAddress
	device.Port = port
	////device.HostUrl = fmt.Sprintf("%s:%d", ipAddress, port)

	return device
}

func (this *NetDevice) GetHostUrl() string {
	return fmt.Sprintf("%s:%d", this.IPAddress, this.Port)
}
package maasclient

import (
	"encoding/json"
	"net"
)

type IPAddress interface {
	IP() net.IP
}

type ipaddress struct {
	ip net.IP
}

func (i *ipaddress) IP() net.IP {
	return i.ip
}

func (i *ipaddress) UnmarshalJSON(data []byte) error {
	des := &struct {
		Ip string `json:"ip"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	i.ip = net.ParseIP(des.Ip)

	return nil
}

func ipStructToInterface(in *ipaddress, client Client) IPAddress {
	return in
}

func ipStructSliceToInterface(in []*ipaddress, client Client) []IPAddress {
	var out []IPAddress
	for _, ip := range in {
		out = append(out, ipStructToInterface(ip, client))
	}
	return out
}

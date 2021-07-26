package maasclient

import "encoding/json"

type VLAN interface {
	VID() int
	MTU() int
	IsDHCPOn() bool
	FabricID() int
	FabricName() string
}

type vLAN struct {
	vid        int
	mtu        int
	dhcpOn     bool
	fabricId   int
	fabricName string
	id         int
	name       string
}

func (v *vLAN) VID() int {
	return v.vid
}

func (v *vLAN) MTU() int {
	return v.mtu
}

func (v *vLAN) IsDHCPOn() bool {
	return v.dhcpOn
}

func (v *vLAN) FabricID() int {
	return v.fabricId
}

func (v *vLAN) FabricName() string {
	return v.fabricName
}

func (v *vLAN) UnmarshalJSON(data []byte) error {
	des := &struct {
		Id       int    `json:"id"`
		VId      int    `json:"vid"`
		Name     string `json:"name"`
		Fabric   string `json:"fabric"`
		FabricID int    `json:"fabric_id"`
		MTU      int    `json:"mtu"`
		DHCPOn   bool   `json:"dhcp_on"`
	}{}

	err := json.Unmarshal(data, des)
	if err != nil {
		return err
	}

	v.id = des.Id
	v.vid = des.VId
	v.name = des.Name
	v.fabricId = des.FabricID
	v.fabricName = des.Fabric
	v.mtu = des.MTU
	v.dhcpOn = des.DHCPOn

	return nil
}

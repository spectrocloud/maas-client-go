package maasclient

type ClientSetInterface interface {
	BootResources() BootResources
	DNSResources() DNSResources
	Domains() Domains
	Machines() Machines
	RackControllers() RackControllers
	ResourcePools() ResourcePools
	Spaces() Spaces
	Users() Users
	Zones() Zones
}

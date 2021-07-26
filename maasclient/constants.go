package maasclient

type FilterType string

const (
	// parametes
	FQDNKey           = "fqdn"
	DomainKey         = "domain"
	NameKey           = "name"
	ArchitectureKey = "architecture"
	SHA256Key        = "sha256"
	SizeKey          = "size"
	ZoneKey          = "zone"
	SystemIDKey      = "system_id"
	CPUCountKey      = "cpu_count"
	MemoryKey        = "mem"
	PoolLabel        = "pool"
	OSSystemKey      = "osystem"
	UserDataKey      = "user_data"
	DistroSeriesKey  = "distro_series"
	TitleKey       = "title"
	ContentKey     = "content"
	FileTypeKey    = "filetype"
	RRTypeKey      = "rrtype"
	AllKey         = "all"
	AddressTTLKey  = "address_ttl"
	IPAddressesKey = "ip_addresses"
	IDKey          = "id"
	EraseKey       = "erase"
	QuickEraseKey  = "quick_erase"
	SecureEraseKey = "secure_erase"
	ForceKey       = "force"
	CommentKey     = "comment"
	SwapSizeKey    = "swap_size"
	HostnameKey    = "hostname"
	TrueKey        = "true"

	// Resource operations
	Operation                 = "op"
	OperationDeploy           = "deploy"
	OperationWhoAmI           = "whoami"
	OperationImportBootImages = "import_boot_images"
	OperationReleaseMachine   = "release"
	OperationAllocate         = "allocate"
)

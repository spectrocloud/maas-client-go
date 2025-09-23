/*
Copyright 2021 Spectro Cloud

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package maasclient

type FilterType string

const (
	// parametes
	FQDNKey         = "fqdn"
	DomainKey       = "domain"
	NameKey         = "name"
	ArchitectureKey = "architecture"
	SHA256Key       = "sha256"
	SizeKey         = "size"
	BaseImageKey    = "base_image"
	ZoneKey         = "zone"
	TagKey          = "tags"
	SystemIDKey     = "system_id"
	CPUCountKey     = "cpu_count"
	MemoryKey       = "mem"
	PoolLabel       = "pool"
	NotPodKey       = "not_pod"
	NotPodTypeKey   = "not_pod_type"
	OSSystemKey     = "osystem"
	UserDataKey     = "user_data"
	DistroSeriesKey = "distro_series"
	TitleKey        = "title"
	ContentKey      = "content"
	FileTypeKey     = "filetype"
	RRTypeKey       = "rrtype"
	AllKey          = "all"
	AddressTTLKey   = "address_ttl"
	IPAddressesKey  = "ip_addresses"
	IDKey           = "id"
	EraseKey        = "erase"
	QuickEraseKey   = "quick_erase"
	SecureEraseKey  = "secure_erase"
	ForceKey        = "force"
	CommentKey      = "comment"
	SwapSizeKey     = "swap_size"
	HostnameKey     = "hostname"
	TrueKey         = "true"
	SubnetKey       = "subnet"
	IPAddressKey    = "ip_address"
	ModeKey         = "mode"
	LinkIDKey       = "id"
	ParentKey       = "parent"

	// Network interface modes
	ModeDHCP   = "dhcp"
	ModeStatic = "static"
	ModeLinkUp = "link_up"

	// Resource operations
	Operation                 = "op"
	OperationDeploy           = "deploy"
	OperationWhoAmI           = "whoami"
	OperationImportBootImages = "import_boot_images"
	OperationReleaseMachine   = "release"
	OperationAllocate         = "allocate"
	OperationLinkSubnet       = "link_subnet"
	OperationUnlinkSubnet     = "unlink_subnet"
	OperationCreateBridge     = "create_bridge"
	OperationReleaseIPAddress = "release"
)

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

import (
	"context"
	"testing"
)

// Step 1: Test listing network interfaces for a selected machine
func TestStep1_ListNetworkInterfaces_RealMAAS(t *testing.T) {
	// Real MAAS server configuration
	maasURL := "<MAAS_URL>"
	apiKey := "<API_KEY>"
	systemID := "<SYSTEM_ID>"

	t.Logf("Step 1: Testing network interface listing for machine %s", systemID)

	client := NewAuthenticatedClientSet(maasURL, apiKey)

	// Test getting network interfaces
	interfaces, err := client.NetworkInterfaces().Get(context.Background(), systemID)
	if err != nil {
		t.Fatalf("Failed to get network interfaces: %v", err)
	}

	t.Logf("✓ Successfully retrieved %d network interfaces", len(interfaces))

	if len(interfaces) == 0 {
		t.Log("No interfaces found for this machine")
		return
	}

	// Display detailed information about each interface
	for i, iface := range interfaces {
		t.Logf("Interface %d:", i+1)
		t.Logf("  - ID: %s", iface.ID())
		t.Logf("  - Name: %s", iface.Name())
		t.Logf("  - Type: %s", iface.Type())
		t.Logf("  - Enabled: %v", iface.Enabled())
		t.Logf("  - MAC Address: %s", iface.MACAddress())

		// Check VLAN info
		vlan := iface.VLAN()
		if vlan != nil {
			t.Logf("  - VLAN: <exists>")
			// Try each method separately to identify which one fails
			defer func() {
				if r := recover(); r != nil {
					t.Logf("  - VLAN: <error accessing VLAN methods: %v>", r)
				}
			}()
			vid := vlan.VID()
			fabricID := vlan.FabricID()
			mtu := vlan.MTU()
			dhcp := vlan.IsDHCPOn()
			t.Logf("  - VLAN: VID=%d, FabricID=%d, MTU=%d, DHCP=%v", vid, fabricID, mtu, dhcp)
		} else {
			t.Logf("  - VLAN: <not configured>")
		}

		// Check links (IP configurations)
		links := iface.Links()
		t.Logf("  - Links: %d configured", len(links))
		for j, link := range links {
			t.Logf("    Link %d: ID=%s, Mode=%s, IP=%s",
				j+1, link.ID(), link.Mode(), link.IPAddress().String())
			if link.Subnet() != nil {
				t.Logf("      Subnet: ID=%d, Name=%s",
					link.Subnet().ID(), link.Subnet().Name())
			}
		}
	}

	t.Log("✓ Step 1 passed: Successfully listed and verified network interfaces")
}

// Test case 2: Boot interface has no links but has children (bridge scenario)
func TestBootInterfaceWithBridge_SetStaticIP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real MAAS server test in short mode")
	}

	maasURL := "<MAAS_URL>"
	apiKey := "<API_KEY>"
	systemID := "<SYSTEM_ID>"
	targetIP := "<TARGET_IP>"

	t.Log("=== Test Case 2: Boot interface with bridge (no direct links) ===")

	client := NewAuthenticatedClientSet(maasURL, apiKey)

	// Set static IP on boot interface
	err := client.NetworkInterfaces().SetBootInterfaceStaticIP(context.Background(), systemID, targetIP)
	if err != nil {
		t.Fatalf("Failed to set static IP: %v", err)
	}

	t.Log("✓ Case 2 passed: Successfully updated bridge interface to static IP")
}

// Test case 1: Boot interface has direct links - can unlink and relink directly
func TestBootInterfaceWithDirectLinks_SetStaticIP(t *testing.T) {
	maasURL := "<MAAS_URL>"
	apiKey := "<API_KEY>"
	systemID := "<SYSTEM_ID>"
	targetIP := "<TARGET_IP>"

	t.Log("=== Test Case 1: Boot interface with direct links ===")

	client := NewAuthenticatedClientSet(maasURL, apiKey)

	err := client.NetworkInterfaces().SetBootInterfaceStaticIP(context.Background(), systemID, targetIP)
	if err != nil {
		t.Fatalf("Failed to set static IP on boot interface: %v", err)
	}

	t.Log("✓ Case 1 passed: Successfully updated boot interface with direct links to static IP")
}

// Test case: Machine extraction functions with real MAAS data
func TestMachine_ExtractionFunctions_RealMAAS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real MAAS server test in short mode")
	}

	maasURL := "<MAAS_URL>"
	apiKey := "<API_KEY>"
	systemID := "<SYSTEM_ID>"

	t.Log("=== Test Case: Machine Extraction Functions ===")

	client := NewAuthenticatedClientSet(maasURL, apiKey)

	// Get machine details
	machineClient := client.Machines().Machine(systemID)
	machine, err := machineClient.Get(context.Background())
	if err != nil {
		t.Fatalf("Failed to get machine details: %v", err)
	}

	// Test ResourcePoolName function
	poolName := machine.ResourcePoolName()
	t.Logf("Resource Pool Name: '%s'", poolName)
	if poolName == "" {
		t.Log("Warning: Resource pool name is empty - machine may not have a pool assigned")
	}

	// Test ZoneName function
	zoneName := machine.ZoneName()
	t.Logf("Zone Name: '%s'", zoneName)
	if zoneName == "" {
		t.Error("Zone name should not be empty - all machines should have a zone")
	}

	// Test BootInterfaceName function
	bootInterfaceName := machine.BootInterfaceName()
	t.Logf("Boot Interface Name: '%s'", bootInterfaceName)
	if bootInterfaceName == "" {
		t.Error("Boot interface name should not be empty - all machines should have a boot interface")
	}

	// Test PowerType function
	powerType := machine.PowerType()
	t.Logf("Power Type: %s", powerType)
	if powerType == "" {
		t.Error("Power type should not be empty - all machines should have a power type")
	}

	// Additional logging for verification
	t.Logf("Machine Details:")
	t.Logf("  - System ID: %s", machine.SystemID())
	t.Logf("  - Hostname: %s", machine.Hostname())
	t.Logf("  - FQDN: %s", machine.FQDN())
	t.Logf("  - State: %s", machine.State())
	t.Logf("  - Power State: %s", machine.PowerState())
	t.Logf("  - Boot Interface ID: %s", machine.BootInterfaceID())
	t.Logf("  - Boot Interface Type: %s", machine.GetBootInterfaceType())

	t.Log("✓ Machine extraction functions test completed")
}

// Test case: Register a new LXD VM host
func TestVMHosts_RegisterLXDHost_RealMAAS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real MAAS server test in short mode")
	}

	maasURL := "<MAAS_URL>"
	apiKey := "<API_KEY>"

	// LXD host parameters - adjust these for your environment
	lxdHostIP := "<LXD_HOST_IP>"       // IP address of the LXD host
	lxdHostName := "<LXD_HOST_NAME>" // Name for the VM host
	zoneName := "<ZONE_NAME>"                  // Zone to place the host in
	poolName := "<POOL_NAME>"            

	// t.Log("=== Test Case: Register LXD VM Host ===")

	client := NewAuthenticatedClientSet(maasURL, apiKey)

	// Prepare parameters for LXD host registration
	params := ParamsBuilder()
	params.Set("type", "lxd")
	params.Set("power_address", lxdHostIP)
	params.Set("name", lxdHostName)
	params.Set("zone", zoneName)
	params.Set("pool", poolName)
	params.Set("password", "capmaasbm-az3-85")

	t.Logf("Registering LXD host:")
	t.Logf("  - Name: %s", lxdHostName)
	t.Logf("  - IP Address: %s", lxdHostIP)
	t.Logf("  - Zone: %s", zoneName)
	t.Logf("  - Pool: %s", poolName)

	// Register the VM host
	vmHost, err := client.VMHosts().Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to register LXD host: %v", err)
	}

	t.Logf("✅ Successfully registered LXD host!")
	t.Logf("VM Host Details:")
	t.Logf("  - System ID: %s", vmHost.SystemID())
	t.Logf("  - Name: %s", vmHost.Name())
	t.Logf("  - Type: %s", vmHost.Type())
	t.Logf("  - Host System ID: %s", vmHost.HostSystemID())

	// Get detailed information about the newly registered host
	detailedHost, err := vmHost.Get(context.Background())
	if err != nil {
		t.Logf("Failed to get detailed host information: %v", err)
	} else {
		t.Logf("Resources:")
		t.Logf("  - CPU Cores: %d total, %d available",
			detailedHost.TotalCores(), detailedHost.AvailableCores())
		t.Logf("  - Memory: %d MB total, %d MB available",
			detailedHost.TotalMemory(), detailedHost.AvailableMemory())

		// Check zone and resource pool assignment
		if detailedHost.Zone() != nil {
			t.Logf("  - Zone: %s (ID: %d)", detailedHost.Zone().Name(), detailedHost.Zone().ID())
		}
		if detailedHost.ResourcePool() != nil {
			t.Logf("  - Resource Pool: %s (ID: %d)", detailedHost.ResourcePool().Name(), detailedHost.ResourcePool().ID())
		}
	}

	t.Log("✓ LXD host registration test completed")
}

// Test case: Retrieve and display specific LXD host information
func TestVMHosts_LXD_HostInfo_RealMAAS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real MAAS server test in short mode")
	}

	maasURL := "<MAAS_URL>"
	apiKey := "<API_KEY>"
	lxdHostID := "<LXD_HOST_ID>" // Replace with actual LXD host system ID

	t.Log("=== Test Case: Single LXD Host Information Retrieval ===")

	client := NewAuthenticatedClientSet(maasURL, apiKey)

	// Get specific VM host by system ID
	vmHost := client.VMHosts().VMHost(lxdHostID)
	detailedHost, err := vmHost.Get(context.Background())
	if err != nil {
		t.Fatalf("Failed to get LXD host %s: %v", lxdHostID, err)
	}

	t.Logf("--- LXD Host Information ---")

	// Basic information
	t.Logf("System ID: %s", detailedHost.SystemID())
	t.Logf("Name: %s", detailedHost.Name())
	t.Logf("Type: %s", detailedHost.Type())
	t.Logf("Power Address: %s", detailedHost.PowerAddress())
	t.Logf("Host System ID: %s", detailedHost.HostSystemID())

	// Get the host machine's boot interface IP address using the host system ID
	hostSystemID := detailedHost.HostSystemID()
	if hostSystemID != "" {
		t.Logf("Looking up host machine with system ID: %s", hostSystemID)
		hostMachine := client.Machines().Machine(hostSystemID)
		hostMachineDetails, err := hostMachine.Get(context.Background())
		if err != nil {
			t.Logf("Failed to get host machine details: %v", err)
		} else {
			t.Logf("Host Machine FQDN: %s", hostMachineDetails.FQDN())
			t.Logf("Host Machine Boot Interface Name: %s", hostMachineDetails.BootInterfaceName())

			// Check if the bare metal host machine is healthy (powered on and deployed)
			powerState := hostMachineDetails.PowerState()
			machineState := hostMachineDetails.State()
			powerType := hostMachineDetails.PowerType()

			t.Logf("Host Machine Power State: %s", powerState)
			t.Logf("Host Machine State: %s", machineState)
			t.Logf("Host Machine Power Type: %s", powerType)

			isHealthy := powerState == "on" && machineState == "Deployed"
			if isHealthy {
				t.Logf("✅ LXD Host Machine is HEALTHY (powered on and deployed)")
			} else {
				t.Logf("❌ LXD Host Machine is UNHEALTHY - Power: %s, State: %s", powerState, machineState)
			}

			// Get the boot interface IP specifically
			bootInterfaceName := hostMachineDetails.BootInterfaceName()
			if bootInterfaceName != "" {
				// Get network interfaces for the host machine
				interfaces, err := client.NetworkInterfaces().Get(context.Background(), hostSystemID)
				if err != nil {
					t.Logf("Failed to get network interfaces for host machine: %v", err)
				} else {
					// Find the boot interface and get its IP
					var bootInterface NetworkInterface
					for _, iface := range interfaces {
						if iface.Name() == bootInterfaceName {
							bootInterface = iface
							break
						}
					}

					if bootInterface != nil {
						t.Logf("Found boot interface: %s", bootInterface.Name())

						// Check if boot interface has direct IP links
						links := bootInterface.Links()
						if len(links) > 0 {
							for _, link := range links {
								if link.IPAddress() != nil {
									t.Logf("Boot Interface IP Address: %s", link.IPAddress().String())
									// This is the LXD host IP address
									break
								}
							}
						} else {
							t.Logf("Boot interface has no direct IP links")

							// For LXD hosts, check if boot interface has children (bridge setup)
							// The children (like br-enp2s0f0) will have the actual IP
							children := bootInterface.Children()
							if len(children) > 0 {
								t.Logf("Boot interface has %d children: %v", len(children), children)

								// Look for IP addresses in child interfaces
								for _, childName := range children {
									for _, iface := range interfaces {
										if iface.Name() == childName {
											t.Logf("Checking child interface: %s", childName)
											childLinks := iface.Links()
											for _, link := range childLinks {
												if link.IPAddress() != nil {
													t.Logf("Child Interface (%s) IP Address: %s", childName, link.IPAddress().String())
													t.Logf("LXD Host IP Address: %s", link.IPAddress().String())
													goto found_ip
												}
											}
											break
										}
									}
								}
							} else {
								t.Logf("Boot interface has no children")
							}
						}
					} else {
						t.Logf("Boot interface %s not found", bootInterfaceName)
					}
				found_ip:
				}
			} else {
				t.Logf("No boot interface name found for host machine")
			}
		}
	} else {
		t.Logf("No host system ID available")
	}

	// Zone and Resource Pool information
	if detailedHost.Zone() != nil {
		t.Logf("Zone: %s (ID: %d)", detailedHost.Zone().Name(), detailedHost.Zone().ID())
	} else {
		t.Logf("Zone: <not assigned>")
	}

	if detailedHost.ResourcePool() != nil {
		t.Logf("Resource Pool: %s (ID: %d)", detailedHost.ResourcePool().Name(), detailedHost.ResourcePool().ID())
	} else {
		t.Logf("Resource Pool: <not assigned>")
	}

	// Resource utilization
	t.Logf("Resources:")
	t.Logf("  - CPU Cores: %d total, %d used, %d available",
		detailedHost.TotalCores(), detailedHost.UsedCores(), detailedHost.AvailableCores())
	t.Logf("  - Memory: %d MB total, %d MB used, %d MB available",
		detailedHost.TotalMemory(), detailedHost.UsedMemory(), detailedHost.AvailableMemory())

	// Storage pools
	storagePools := detailedHost.StoragePools()
	if len(storagePools) > 0 {
		t.Logf("Storage Pools:")
		for j, pool := range storagePools {
			t.Logf("  %d. Name: %s, Driver: %s, Total: %d, Used: %d",
				j+1, pool.Name, pool.Driver, pool.Total, pool.Used)
		}
	} else {
		t.Logf("Storage Pools: <none>")
	}

	// Get machines on this VM host
	vmHostMachines := detailedHost.Machines()
	machines, err := vmHostMachines.List(context.Background())
	if err != nil {
		t.Logf("Failed to get machines for VM host %s: %v", detailedHost.SystemID(), err)
	} else {
		t.Logf("Virtual Machines: %d", len(machines))
		for j, machine := range machines {
			t.Logf("  %d. %s (%s) - State: %s",
				j+1, machine.Hostname(), machine.SystemID(), machine.State())
		}
	}

	t.Log("✓ LXD host information retrieval test completed")

	// Summary
	t.Logf("=== LXD Host Summary ===")
	t.Logf("LXD Host: %s (ID: %s)", detailedHost.Name(), detailedHost.SystemID())
	t.Logf("Host Machine: %s", hostSystemID)
	if hostSystemID != "" {
		hostMachine := client.Machines().Machine(hostSystemID)
		if hostMachineDetails, err := hostMachine.Get(context.Background()); err == nil {
			isHealthy := hostMachineDetails.PowerState() == "on" && hostMachineDetails.State() == "Deployed"
			healthStatus := "HEALTHY"
			if !isHealthy {
				healthStatus = "UNHEALTHY"
			}
			t.Logf("Health Status: %s", healthStatus)
		}
	}
}


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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const nodeDevicesResponse = `[
  {
    "id": 159,
    "bus": 1,
    "bus_name": "PCIE",
    "hardware_type": 5,
    "hardware_type_name": "gpu",
    "vendor_id": "10de",
    "product_id": "1eb8",
    "vendor_name": "NVIDIA Corporation",
    "product_name": "TU104GL [Tesla T4]",
    "commissioning_driver": "pci",
    "bus_number": 0,
    "device_number": 30,
    "pci_address": "0000:00:1e.0",
    "numa_node": 0,
    "physical_blockdevice": null,
    "physical_interface": null,
    "system_id": "abc123",
    "resource_uri": "/MAAS/api/2.0/nodes/abc123/devices/159/"
  },
  {
    "id": 42,
    "bus": 1,
    "bus_name": "PCIE",
    "hardware_type": 4,
    "hardware_type_name": "network",
    "vendor_id": "8086",
    "product_id": "37d2",
    "vendor_name": "Intel Corporation",
    "product_name": "Ethernet Connection X722",
    "commissioning_driver": "i40e",
    "bus_number": 0,
    "device_number": 25,
    "pci_address": "0000:00:19.0",
    "numa_node": 1,
    "physical_blockdevice": null,
    "physical_interface": null,
    "system_id": "abc123",
    "resource_uri": "/MAAS/api/2.0/nodes/abc123/devices/42/"
  }
]`

func TestNodeDevicesList(t *testing.T) {
	ctx := context.Background()

	t.Run("lists devices for a machine", func(t *testing.T) {
		var gotMethod, gotPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, nodeDevicesResponse)
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		devices, err := c.NodeDevices().List(ctx, "abc123", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/api/2.0/nodes/abc123/devices/", gotPath)
		assert.Len(t, devices, 2)

		gpu := devices[0]
		assert.Equal(t, "159", gpu.ID())
		assert.Equal(t, "PCIE", gpu.BusName())
		assert.Equal(t, "gpu", gpu.HardwareTypeName())
		assert.Equal(t, "10de", gpu.VendorID())
		assert.Equal(t, "1eb8", gpu.ProductID())
		assert.Equal(t, "NVIDIA Corporation", gpu.VendorName())
		assert.Equal(t, "TU104GL [Tesla T4]", gpu.ProductName())
		assert.Equal(t, "pci", gpu.CommissioningDriver())
		assert.Equal(t, 0, gpu.BusNumber())
		assert.Equal(t, 30, gpu.DeviceNumber())
		assert.Equal(t, "0000:00:1e.0", gpu.PCIAddress())
		assert.Equal(t, 0, gpu.NUMANode())
		assert.Equal(t, "abc123", gpu.SystemID())
		assert.Equal(t, "/MAAS/api/2.0/nodes/abc123/devices/159/", gpu.ResourceUri())

		nic := devices[1]
		assert.Equal(t, "42", nic.ID())
		assert.Equal(t, "network", nic.HardwareTypeName())
		assert.Equal(t, 1, nic.NUMANode())
	})

	t.Run("passes filter params as query string", func(t *testing.T) {
		var gotQuery string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.Query().Get(HardwareTypeKey)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "[]")
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		params := ParamsBuilder().Set(HardwareTypeKey, HardwareTypeGPU)
		devices, err := c.NodeDevices().List(ctx, "abc123", params)
		assert.Nil(t, err)
		assert.Equal(t, HardwareTypeGPU, gotQuery)
		assert.Empty(t, devices)
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Not Found")
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		devices, err := c.NodeDevices().List(ctx, "missing", nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "status: 404")
		assert.Nil(t, devices)
	})
}

// TestNodeDevicesIntegration runs against a real MAAS. It follows the same
// conventions as the other integration tests in this package: set
// MAAS_ENDPOINT and MAAS_API_KEY, and replace the placeholder system ID.
func TestNodeDevicesIntegration(t *testing.T) {
	endPoint := os.Getenv("MAAS_ENDPOINT")
	apiKey := os.Getenv("MAAS_API_KEY")
	if endPoint == "" || apiKey == "" {
		t.Skip("MAAS_ENDPOINT and MAAS_API_KEY must be set for integration tests")
	}
	c := NewAuthenticatedClientSet(endPoint, apiKey)

	ctx := context.Background()

	// TODO: Replace with an actual machine system ID
	systemID := "REPLACE_WITH_MACHINE_SYSTEM_ID_1"
	if systemID == "REPLACE_WITH_MACHINE_SYSTEM_ID_1" {
		t.Skip("Please replace placeholder machine system ID with an actual machine")
		return
	}

	t.Run("list all devices", func(t *testing.T) {
		devices, err := c.NodeDevices().List(ctx, systemID, nil)
		assert.Nil(t, err)
		for _, d := range devices {
			fmt.Printf("%s %s %s %s\n", d.ID(), d.HardwareTypeName(), d.VendorName(), d.ProductName())
		}
	})

	t.Run("list gpu devices only", func(t *testing.T) {
		params := ParamsBuilder().Set(HardwareTypeKey, HardwareTypeGPU)
		devices, err := c.NodeDevices().List(ctx, systemID, params)
		assert.Nil(t, err)
		for _, d := range devices {
			assert.Equal(t, HardwareTypeGPU, d.HardwareTypeName())
		}
	})
}

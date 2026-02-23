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
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// roundTripperFunc allows us to stub http.RoundTripper inline
type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func newMockClientSet(t *testing.T, handler roundTripperFunc) ClientSetInterface {
	transport := &http.Client{Transport: handler}
	return NewAuthenticatedClientSet(
		"http://example",
		"consumer:key:secret",
		func(cs *authenticatedClientSet) { cs.client.httpClient = transport },
	)
}

func TestVMHosts_List(t *testing.T) {
	listBody := `[
		{"id": 1, "name": "host-1", "type": "lxd", "power_address": "https://10.0.0.1:8443",
		 "zone": {"id": 11, "name": "az1"}, "pool": {"id": 21, "name": "pool-a"},
		 "total": {"cores": 32, "memory": 65536},
		 "used": {"cores": 8, "memory": 16384},
		 "available": {"cores": 24, "memory": 49152},
		 "capabilities": ["instances"], "projects": ["default"],
		 "storage_pools": [{"name": "default", "driver": "zfs"}]}
	]`

	client := newMockClientSet(t, func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/api/2.0/vm-hosts/" {
			return jsonResponse(200, listBody), nil
		}
		return jsonResponse(404, `{"error":"not found"}`), nil
	})

	ctx := context.Background()
	hosts, err := client.VMHosts().List(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, hosts, 1)
	assert.Equal(t, "1", hosts[0].SystemID())
	assert.Equal(t, "host-1", hosts[0].Name())
	assert.Equal(t, "lxd", hosts[0].Type())
	assert.Equal(t, "https://10.0.0.1:8443", hosts[0].PowerAddress())
	assert.Equal(t, "az1", hosts[0].Zone().Name())
	assert.Equal(t, "pool-a", hosts[0].ResourcePool().Name())
}

func TestClient_Create(t *testing.T) {
	os.Setenv("MAAS_ENDPOINT", "http://10.11.130.11:5240/MAAS")
	os.Setenv("MAAS_API_KEY", "NfZAfdJTMNs5tKaN6s:pDzCMs6eyDr9qeFvcc:k8QgJbQMBTpKNf57TFdSrsdaga4v3g2x")
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	nodeIP := "10.11.130.70"
	// Create registration parameters
	params1 := ParamsBuilder().
		Set("type", "lxd").
		Set("hostname", nodeIP).
		Set("power_address", fmt.Sprintf("https://%s:8443", nodeIP)).
		Set("name", fmt.Sprintf("lxd-host-%s", nodeIP))
	// Step 1: allocate
	res, err := c.VMHosts().Create(ctx, params1)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Created:", res.SystemID(), res) // State == "Deployed"
}

func TestVMHosts_Create_Get_Update_Delete(t *testing.T) {
	createResp := `{"id": 2, "name": "host-2", "type": "lxd", "power_address": "https://10.0.0.2:8443",
		"zone": {"id": 12, "name": "az2"}, "pool": {"id": 22, "name": "pool-b"},
		"total": {"cores": 48, "memory": 98304},
		"used": {"cores": 0, "memory": 0},
		"available": {"cores": 48, "memory": 98304},
		"capabilities": ["instances"], "projects": ["default"],
		"storage_pools": [{"name": "default", "driver": "zfs"}]}
	`
	getResp := strings.ReplaceAll(createResp, "\"name\": \"host-2\"", "\"name\": \"host-2a\"")
	updateResp := strings.ReplaceAll(createResp, "\"name\": \"host-2\"", "\"name\": \"host-2b\"")

	client := newMockClientSet(t, func(req *http.Request) (*http.Response, error) {
		path := req.URL.Path
		switch req.Method {
		case http.MethodPost:
			if path == "/api/2.0/vm-hosts/" {
				return jsonResponse(200, createResp), nil
			}
		case http.MethodGet:
			if path == "/api/2.0/vm-hosts/2/" {
				return jsonResponse(200, getResp), nil
			}
		case http.MethodPut:
			if path == "/api/2.0/vm-hosts/2/" {
				return jsonResponse(200, updateResp), nil
			}
		case http.MethodDelete:
			if path == "/api/2.0/vm-hosts/2/" {
				return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
			}
		}
		return jsonResponse(404, `{"error":"not found"}`), nil
	})

	ctx := context.Background()
	params := ParamsBuilder().Set("type", "lxd").Set("power_address", "https://10.0.0.2:8443").Set("name", "host-2")
	created, err := client.VMHosts().Create(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, "2", created.SystemID())
	assert.Equal(t, "host-2", created.Name())

	// Get should populate data and normalize system id
	fetched, err := client.VMHosts().VMHost("2").Get(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "2", fetched.SystemID())
	assert.Equal(t, "host-2a", fetched.Name())
	assert.Equal(t, "az2", fetched.Zone().Name())

	// Update should return new data
	updated, err := client.VMHosts().VMHost("2").Update(ctx, ParamsBuilder().Set("name", "host-2b"))
	assert.NoError(t, err)
	assert.Equal(t, "host-2b", updated.Name())

	// Delete should succeed
	derr := client.VMHosts().VMHost("2").Delete(ctx)
	assert.NoError(t, derr)
}

func TestVMHost_ComposeAndMachinesList(t *testing.T) {
	composeMachineID := "abcd12"
	machinesBody := fmt.Sprintf(`[{"system_id": %q}, {"system_id": %q}]`, composeMachineID, "ef3456")

	client := newMockClientSet(t, func(req *http.Request) (*http.Response, error) {
		path := req.URL.Path
		switch req.Method {
		case http.MethodPost:
			if path == "/api/2.0/vm-hosts/9/" {
				// Verify op=compose is present in body
				b, _ := io.ReadAll(req.Body)
				// Ensure request body is application/x-www-form-urlencoded
				vals, _ := url.ParseQuery(string(b))
				if vals.Get("op") != "compose" {
					return jsonResponse(400, `{"error":"missing op=compose"}`), nil
				}
				return jsonResponse(200, fmt.Sprintf(`{"system_id": %q}`, composeMachineID)), nil
			}
		case http.MethodGet:
			if path == "/api/2.0/vm-hosts/9/machines/" {
				return jsonResponse(200, machinesBody), nil
			}
		}
		return jsonResponse(404, `{"error":"not found"}`), nil
	})

	ctx := context.Background()
	vmh := client.VMHosts().VMHost("9")

	// Compose returns a Machine client pointing at the composed system-id
	m, err := vmh.Composer().Compose(ctx, ParamsBuilder().Set("hostname", "vm1").Set("memory", "4096"))
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, composeMachineID, m.SystemID())

	// Machines().List should return machine clients for the VM host
	list, err := vmh.Machines().List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	ids := []string{list[0].SystemID(), list[1].SystemID()}
	assert.ElementsMatch(t, []string{composeMachineID, "ef3456"}, ids)
}

func TestVMHostList(t *testing.T) {

	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	//params1 := ParamsBuilder().
	//	Set("type", "lxd").
	//	Set("hostname", nodeIP).
	//	Set("power_address", fmt.Sprintf("https://%s:8443", nodeIP)).
	//	Set("name", fmt.Sprintf("lxd-host-%s", nodeIP))

	vmhList, _ := c.VMHosts().List(ctx, nil)

	for _, vmh := range vmhList {
		fmt.Print(" Name: ", vmh.Name())
		fmt.Print(" SystemID: ", vmh.SystemID())
		fmt.Print(" HostSystemID: ", vmh.HostSystemID())
		fmt.Println(" PowerAddress: ", vmh.PowerAddress())

		host, _ := vmh.Get(ctx)
		fmt.Println(" host: ", host)
		//fmt.Printf("%v", assert.Len(t, vmh, 2))
	}
}

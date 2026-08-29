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
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetMachine(t *testing.T) {
	c := requireMAASIntegration(t)
	ctx := context.Background()
	systemID := discoverMachineSystemID(t, c, ctx)

	res, err := c.Machines().Machine(systemID).Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, systemID, res.SystemID())
	assert.NotEmpty(t, res.Hostname())
	assert.NotEmpty(t, res.State())
	assert.NotEmpty(t, res.PowerState())
}

func TestClient_AllocateMachine(t *testing.T) {
	requireMachineMutations(t)
	c := requireMAASIntegration(t)
	ctx := context.Background()

	releaseMachine := func(res Machine) {
		if res != nil {
			_, err := res.Releaser().
				WithComment("maas-client-go test release").
				Release(ctx)
			assert.Nil(t, err)
		}
	}

	t.Run("no-options", func(t *testing.T) {
		res, err := c.Machines().Allocator().Allocate(ctx)
		require.NoError(t, err)
		require.NotNil(t, res)

		releaseMachine(res)
	})

	t.Run("bad-options", func(t *testing.T) {
		res, err := c.Machines().
			Allocator().
			WithSystemID("abc").
			Allocate(ctx)

		assert.Error(t, err)
		releaseMachine(res)
	})

	t.Run("with-zone", func(t *testing.T) {
		zone := discoverTestZone(t, c, ctx)

		res, err := c.Machines().Allocator().WithZone(zone).Allocate(ctx)
		require.NoError(t, err)
		require.NotNil(t, res)

		releaseMachine(res)
	})
}

func TestClient_DeployMachine(t *testing.T) {
	requireMachineMutations(t)
	c := requireMAASIntegration(t)
	ctx := context.Background()

	distroSeries := os.Getenv("MAAS_TEST_DISTRO_SERIES")
	if distroSeries == "" {
		t.Skip("MAAS_TEST_DISTRO_SERIES must be set for deploy tests")
	}

	releaseMachine := func(res Machine) {
		if res != nil {
			_, err := res.Releaser().
				WithComment("maas-client-go test release").
				Release(ctx)
			assert.Nil(t, err)
		}
	}

	t.Run("simple", func(t *testing.T) {
		res, err := c.Machines().Allocator().Allocate(ctx)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.NotEmpty(t, res.SystemID())

		res, err = res.Modifier().SetSwapSize(0).Update(ctx)
		require.NoError(t, err)

		_, err = res.Deployer().
			SetOSSystem("custom").
			SetDistroSeries(distroSeries).Deploy(ctx)
		require.NoError(t, err)
		require.NotNil(t, res)

		assert.Equal(t, "custom", res.OSSystem())
		assert.Equal(t, distroSeries, res.DistroSeries())

		time.Sleep(15 * time.Second)

		releaseMachine(res)
	})
}

func TestClient_UpdateMachine(t *testing.T) {
	requireMachineMutations(t)
	c := requireMAASIntegration(t)
	ctx := context.Background()
	systemID := discoverMachineSystemID(t, c, ctx)

	res, err := c.Machines().Machine(systemID).
		Modifier().
		SetSwapSize(10).
		Update(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 10, res.SwapSize())
}

func TestWithTags_AllTagsPresentInParams(t *testing.T) {
	m := &machines{Controller{
		apiPath: "/machines/",
		params:  ParamsBuilder(),
	}}

	tags := []string{"tag1", "tag2", "tag3"}
	m.WithTags(tags)

	got := m.params.Values()[TagKey]
	assert.Len(t, got, 3, "all tags should be present in params")
	assert.ElementsMatch(t, tags, got)
}

func TestWithTags_SingleTag(t *testing.T) {
	m := &machines{Controller{
		apiPath: "/machines/",
		params:  ParamsBuilder(),
	}}

	m.WithTags([]string{"only-tag"})

	got := m.params.Values()[TagKey]
	assert.Len(t, got, 1)
	assert.Equal(t, "only-tag", got[0])
}

func TestWithTags_Empty(t *testing.T) {
	m := &machines{Controller{
		apiPath: "/machines/",
		params:  ParamsBuilder(),
	}}

	m.WithTags([]string{})

	got := m.params.Values()[TagKey]
	assert.Empty(t, got)
}

const machinesListResponse = `[
  {
    "system_id": "abc123",
    "hostname": "node-1",
    "fqdn": "node-1.maas",
    "status_name": "Ready",
    "architecture": "amd64/generic",
    "cpu_count": 8,
    "memory": 16384,
    "zone": {"id": 1, "name": "az1", "description": ""},
    "pool": {"id": 0, "name": "default", "description": ""},
    "tag_names": ["gpu", "team-a"]
  }
]`

// Regression test: List used to ignore its params argument and always send
// the controller's internal (empty) params, so filters such as tags never
// reached the MAAS API.
func TestMachinesList_PassesParams(t *testing.T) {
	ctx := context.Background()

	t.Run("forwards caller params as query string", func(t *testing.T) {
		var gotQuery url.Values
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.Query()
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, machinesListResponse)
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		params := ParamsBuilder().Add(TagKey, "gpu").Add(TagKey, "team-a")
		machines, err := c.Machines().List(ctx, params)
		assert.Nil(t, err)
		assert.Equal(t, []string{"gpu", "team-a"}, gotQuery[TagKey])
		assert.Len(t, machines, 1)
	})

	t.Run("nil params sends no query string", func(t *testing.T) {
		var gotQuery string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, machinesListResponse)
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		machines, err := c.Machines().List(ctx, nil)
		assert.Nil(t, err)
		assert.Empty(t, gotQuery)
		assert.Len(t, machines, 1)
	})
}

func TestMachinesList_MachineFields(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, machinesListResponse)
	}))
	defer server.Close()

	c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

	machines, err := c.Machines().List(ctx, nil)
	assert.Nil(t, err)
	assert.Len(t, machines, 1)

	m := machines[0]
	assert.Equal(t, "abc123", m.SystemID())
	assert.Equal(t, "node-1", m.Hostname())
	assert.Equal(t, "Ready", m.State())
	assert.Equal(t, 8, m.CPUCount())
	assert.Equal(t, 16384, m.Memory())
	assert.Equal(t, "amd64/generic", m.Architecture())
	assert.Equal(t, "az1", m.ZoneName())
	assert.Equal(t, "default", m.ResourcePoolName())
	assert.Equal(t, []string{"gpu", "team-a"}, m.Tags())
}

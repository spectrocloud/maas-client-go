/*
Copyright 2021 Spectrocloud

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
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	code := m.Run()
	os.Exit(code)
}

func TestClient_GetMachine(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()
	res := c.Machines().Machine("e37xxm")
	err := res.Get(ctx)
	//.machine(ctx, "e37xxm")

	assert.Nil(t, err, "expecting nil error")

	assert.NotNil(t, res, "expecting non-nil result")
	assert.NotEmpty(t, res.SystemID())
	assert.NotEmpty(t, res.Hostname())
	assert.Equal(t, res.State(), "Deployed")
	assert.NotEmpty(t, res.PowerState())
	assert.Equal(t, res.Zone().Name(), "az2")

	assert.NotEmpty(t, res.FQDN())
	assert.NotEmpty(t, res.IPAddresses())

	assert.NotEmpty(t, res.OSSystem())
	assert.NotEmpty(t, res.DistroSeries())

	assert.Zero(t, res.SwapSize())

}

func TestClient_AllocateMachine(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	releaseMachine := func(res Machine) {
		if res != nil {
			err := res.Releaser().
				WithComment("releaseaan").
				Release(ctx)
			assert.Nil(t, err)
			assert.NotNil(t, res)
		}
	}

	t.Run("no-options", func(t *testing.T) {
		res, err := c.Machines().Allocator().Allocate(ctx)

		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res)

		releaseMachine(res)
	})

	t.Run("bad-options", func(t *testing.T) {
		res, err := c.Machines().
			Allocator().
			WithSystemID("abc").
			Allocate(ctx)

		assert.NotNil(t, err, "expecting error")

		releaseMachine(res)
	})

	t.Run("with-az", func(t *testing.T) {
		res, err := c.Machines().Allocator().WithZone("az1").Allocate(ctx)

		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res)

		releaseMachine(res)
	})

}

func TestClient_DeployMachine(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	releaseMachine := func(res Machine) {
		if res != nil {
			err := res.Releaser().
				WithComment("releaseaan a").
				Release(ctx)
			assert.Nil(t, err)
		}
	}

	t.Run("simple", func(t *testing.T) {
		res, err := c.Machines().Allocator().Allocate(ctx)
		if err != nil {
			t.Fatal("Machine didn't allocate")
		}
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.SystemID())

		err = res.Deployer().
			SetOSSystem("custom").
			SetDistroSeries("u-1804-0-k-11912-0").Deploy(ctx)
		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res)

		assert.Equal(t, res.OSSystem(), "custom")
		assert.Equal(t, res.DistroSeries(), "u-1804-0-k-11912-0")

		// Give me a few seconds before clenaing up
		time.Sleep(15 * time.Second)

		releaseMachine(res)
	})

}

func TestClient_UpdateMachine(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	res, err := c.Machines().Machine("e37xxm").
		Modifier().
		SetSwapSize(10).
		Update(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.SwapSize(), 10)

}

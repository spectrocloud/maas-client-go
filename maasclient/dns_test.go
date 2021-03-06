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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDNSResources(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("no-options", func(t *testing.T) {
		res, err := c.DNSResources().List(ctx, nil)
		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res, "expecting non-nil result")

		assert.Greater(t, len(res), 0, "expecting non-empty dns_resources")

		assert.NotZero(t, res[0].ID())
		assert.NotEmpty(t, res[0].FQDN())
	})

	t.Run("invalid-search", func(t *testing.T) {
		filters := ParamsBuilder().Add(FQDNKey, "bad-doesntexist.maas")
		res, err := c.DNSResources().List(ctx, filters)
		assert.NotNil(t, err, "expecting nil error")
		assert.Nil(t, res, "expecting non-nil result")
		assert.Empty(t, res)
	})

	t.Run("get maas-1.maas", func(t *testing.T) {
		filters := ParamsBuilder().Add(FQDNKey, "maas-1.maas.sc")
		res, err := c.DNSResources().List(ctx, filters)
		assert.Nil(t, err, "expecting nil error")
		assert.NotEmpty(t, res)
		assert.NotZero(t, res[0].AddressTTL())
		assert.NotEmpty(t, res[0].IPAddresses())
		assert.NotEmpty(t, res[0].IPAddresses()[0].IP())

		// TODO create test DNS

	})

	t.Run("create test-unit-1.maas", func(t *testing.T) {
		res, err := c.DNSResources().
			Builder().
			WithFQDN("test-unit-1.maas.sc").
			WithAddressTTL("10").Create(ctx)
		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res)
		assert.Equal(t, res.FQDN(), "test-unit-1.maas.sc")
		assert.Equal(t, res.AddressTTL(), 10)
		assert.Empty(t, res.IPAddresses())

		err = res.Delete(ctx)
		assert.Nil(t, err, "expecting nil error")

	})

	t.Run("create test-unit-2.maas", func(t *testing.T) {

		//err := c.DNSResources().DNSResource(148).Delete(ctx)
		//assert.Nil(t, err)

		res, err := c.DNSResources().
			Builder().
			WithFQDN("test-unit-2.maas.sc").
			WithAddressTTL("10").Create(ctx)
		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res)
		assert.Equal(t, res.FQDN(), "test-unit-2.maas.sc")
		assert.Equal(t, res.AddressTTL(), 10)
		assert.Empty(t, res.IPAddresses())

		res, err = res.Modifier().
			SetIPAddresses([]string{"1.2.3.4", "5.6.7.8"}).
			Modify(ctx)
		if err != nil {
			t.Fatal("error", err)
		}
		assert.Equal(t, res.FQDN(), "test-unit-2.maas.sc")
		assert.Equal(t, res.AddressTTL(), 10)
		assert.NotEmpty(t, res.IPAddresses())

		res2, err := c.DNSResources().DNSResource(res.ID()).Get(ctx)
		assert.Nil(t, err)
		assert.True(t, len(res2.IPAddresses()) == 2)

		err = res.Delete(ctx)
		assert.Nil(t, err, "expecting nil error")

	})

	//assert.Equal(t, 1, res.Count, "expecting 1 resource")

	//assert.Equal(t, 1, res.PagesCount, "expecting 1 PAGE found")
	//
	//assert.Equal(t, "integration_face_id", res.Faces[0].FaceID, "expecting correct face_id")
	//assert.NotEmpty(t, res.Faces[0].FaceToken, "expecting non-empty face_token")
	//assert.Greater(t, len(res.Faces[0].FaceImages), 0, "expecting non-empty face_images")
}

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

	"github.com/stretchr/testify/assert"
)

func TestGetDNSResources(t *testing.T) {
	c := requireMAASIntegration(t)
	ctx := context.Background()
	domain := testDNSDomain(t, c, ctx)

	t.Run("no-options", func(t *testing.T) {
		res, err := c.DNSResources().List(ctx, nil)
		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res, "expecting non-nil result")

		assert.Greater(t, len(res), 0, "expecting non-empty dns_resources")

		assert.NotZero(t, res[0].ID())
		assert.NotEmpty(t, res[0].FQDN())
	})

	t.Run("search nonexistent fqdn", func(t *testing.T) {
		nonexistent := dnsFQDN(uniqueDNSResourceName("nonexistent"), domain)
		filters := ParamsBuilder().Add(FQDNKey, nonexistent)
		res, err := c.DNSResources().List(ctx, filters)
		assert.Nil(t, err)
		assert.Empty(t, res)
	})

	t.Run("get by fqdn", func(t *testing.T) {
		existingFQDN := discoverDNSResourceFQDN(t, c, ctx)

		filters := ParamsBuilder().Add(FQDNKey, existingFQDN)
		res, err := c.DNSResources().List(ctx, filters)
		assert.Nil(t, err, "expecting nil error")
		assert.NotEmpty(t, res)
		assert.Equal(t, existingFQDN, res[0].FQDN())
	})

	t.Run("create and delete", func(t *testing.T) {
		testFQDN := dnsFQDN(uniqueDNSResourceName("maas-client-go"), domain)

		res, err := c.DNSResources().
			Builder().
			WithFQDN(testFQDN).
			WithAddressTTL("10").Create(ctx)
		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res)
		assert.Equal(t, testFQDN, res.FQDN())
		assert.Equal(t, 10, res.AddressTTL())
		assert.Empty(t, res.IPAddresses())

		err = res.Delete(ctx)
		assert.Nil(t, err, "expecting nil error")
	})

	t.Run("create modify and delete", func(t *testing.T) {
		testFQDN := dnsFQDN(uniqueDNSResourceName("maas-client-go"), domain)

		res, err := c.DNSResources().
			Builder().
			WithFQDN(testFQDN).
			WithAddressTTL("10").Create(ctx)
		assert.Nil(t, err, "expecting nil error")
		assert.NotNil(t, res)
		assert.Equal(t, testFQDN, res.FQDN())
		assert.Equal(t, 10, res.AddressTTL())
		assert.Empty(t, res.IPAddresses())

		res, err = res.Modifier().
			SetIPAddresses([]string{"1.2.3.4", "5.6.7.8"}).
			Modify(ctx)
		if err != nil {
			t.Fatal("error", err)
		}
		assert.Equal(t, testFQDN, res.FQDN())
		assert.Equal(t, 10, res.AddressTTL())
		assert.NotEmpty(t, res.IPAddresses())

		res2, err := c.DNSResources().DNSResource(res.ID()).Get(ctx)
		assert.Nil(t, err)
		assert.Len(t, res2.IPAddresses(), 2)

		err = res.Delete(ctx)
		assert.Nil(t, err, "expecting nil error")
	})
}

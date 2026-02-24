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

func TestVMHost_Tags(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("list vmhosts and verify tags field", func(t *testing.T) {
		hosts, err := c.VMHosts().List(ctx, nil)
		assert.Nil(t, err)
		assert.NotNil(t, hosts)

		if len(hosts) == 0 {
			t.Skip("No VM hosts available to test")
		}

		// Verify Tags() returns a slice (not nil) for each host
		for _, host := range hosts {
			tagsTest := host.Tags()
			assert.NotNil(t, tagsTest, "Tags() should return empty slice, not nil for host %s", host.Name())
		}
	})

	t.Run("get vmhost and verify tags field", func(t *testing.T) {
		hosts, err := c.VMHosts().List(ctx, nil)
		assert.Nil(t, err)

		if len(hosts) == 0 {
			t.Skip("No VM hosts available to test")
		}

		// Get the first host by ID and verify tags
		host, err := c.VMHosts().VMHost(hosts[0].SystemID()).Get(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, host)

		tagsTest := host.Tags()
		assert.NotNil(t, tagsTest, "Tags() should return empty slice, not nil")
	})
}

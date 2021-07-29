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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestUsers(t *testing.T) {
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("list users", func(t *testing.T) {
		res, err := c.Users().List(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res)
	})

	t.Run("whoami", func(t *testing.T) {
		res, err := c.Users().WhoAmI(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res)
		assert.Equal(t, res.UserName(), "dev")
	})

}

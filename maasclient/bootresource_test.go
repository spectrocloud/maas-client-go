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
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBootResources(t *testing.T) {
	c := requireMAASIntegration(t)
	ctx := context.Background()

	t.Run("list-all", func(t *testing.T) {
		list, err := c.BootResources().List(ctx, nil)
		assert.Nil(t, err, "expecting nil error")
		assert.NotEmpty(t, list)
	})

	t.Run("get-by-id", func(t *testing.T) {
		list, err := c.BootResources().List(ctx, nil)
		assert.Nil(t, err)
		if len(list) == 0 {
			t.Skip("no boot resources in this MAAS environment")
		}

		id := list[0].ID()
		if envID := os.Getenv("MAAS_TEST_BOOT_RESOURCE_ID"); envID != "" {
			parsed, parseErr := strconv.Atoi(envID)
			assert.Nil(t, parseErr)
			id = parsed
		}

		res, err := c.BootResources().BootResource(id).Get(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})

	t.Run("import image", func(t *testing.T) {
		filePath := os.Getenv("MAAS_TEST_BOOT_RESOURCE_FILE")
		if filePath == "" {
			t.Skip("MAAS_TEST_BOOT_RESOURCE_FILE is not set")
		}

		info, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				t.Skipf("boot resource file not found: %s", filePath)
			}
			t.Fatalf("stat boot resource file: %v", err)
		}

		hash, err := sha256File(filePath)
		assert.Nil(t, err)

		name := os.Getenv("MAAS_TEST_BOOT_RESOURCE_NAME")
		if name == "" {
			name = uniqueDNSResourceName("maas-client-go-boot")
		}

		architecture := os.Getenv("MAAS_TEST_BOOT_RESOURCE_ARCH")
		if architecture == "" {
			architecture = "amd64/generic"
		}

		res, err := c.BootResources().Builder(
			name,
			architecture,
			hash,
			filePath,
			int(info.Size()),
		).Create(ctx)
		assert.Nil(t, err)
		if res == nil {
			t.Fatal("expected boot resource after create")
		}
		t.Cleanup(func() {
			_ = res.Delete(ctx)
		})

		err = res.Upload(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

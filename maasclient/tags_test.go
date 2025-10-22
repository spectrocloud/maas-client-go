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

// Test Instructions:
// 1. Replace the placeholder machine system IDs in the test cases with actual unlocked machines
// 2. Ensure the MAAS_ENDPOINT and MAAS_API_KEY environment variables are set
// 3. Run: go test -v ./maasclient -run TestTags
//
// To find unlocked machines, you can:
// - Check MAAS web UI for machines that are not locked
// - Use MAAS API: GET /MAAS/api/2.0/machines/ and look for machines with "locked": false
// - Use the list_tags test to see available machines

package maasclient

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	endPoint := os.Getenv("MAAS_ENDPOINT")
	apiKey := os.Getenv("MAAS_API_KEY")
	c := NewAuthenticatedClientSet(endPoint, apiKey)

	// Uncomment below for Unit Testing purposes.
	//os.Setenv("MAAS_ENDPOINT", "<YOUR_MAAS_ENDPOINT>")
	//os.Setenv("MAAS_API_KEY", "<YOUR_MAAS_API_KEY>")
	//c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	t.Run("list tags", func(t *testing.T) {
		res, err := c.Tags().List(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		for _, eachTag := range res {
			fmt.Println(eachTag.Name())
		}
	})

	t.Run("create tag", func(t *testing.T) {
		err := c.Tags().Create(ctx, "testCase-tag-1")
		assert.Nil(t, err)

		err = c.Tags().Create(ctx, "testCase-tag-2")
		assert.Nil(t, err)

		res, err := c.Tags().List(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		for _, eachTag := range res {
			fmt.Println(eachTag.Name())
		}
	})

	t.Run("assign tag to machines", func(t *testing.T) {
		// First, create a test tag
		tagName := "test-assign-unassign-tag"
		err := c.Tags().Create(ctx, tagName)
		assert.Nil(t, err, "Failed to create tag")

		// TODO: Replace with actual machine system IDs you want to test with
		// These should be unlocked machines that can have tags assigned
		systemID := "REPLACE_WITH_MACHINE_SYSTEM_ID_1"

		// Skip test if placeholder values are still present
		if systemID == "REPLACE_WITH_MACHINE_SYSTEM_ID_1" {
			t.Skip("Please replace placeholder machine system IDs with actual unlocked machines")
			return
		}

		// Assign the tag to the machines
		err = c.Tags().Assign(ctx, tagName, systemID)
		assert.Nil(t, err, "Failed to assign tag")
		fmt.Printf("Successfully assigned tag '%s' to machine: %v\n", tagName, systemID)

		// Wait a bit for the assignment to propagate
		time.Sleep(2 * time.Second)

		machine := c.Machines().Machine(systemID)
		detailedMachine, err := machine.Get(ctx)
		assert.Nil(t, err, "Failed to get machine details for %s", systemID)
		tags := detailedMachine.Tags()
		assert.Contains(t, tags, tagName,
			"Tag '%s' not found on machine %s. Machine tags: %v", tagName, systemID, tags)
		fmt.Printf("✅ Verified tag '%s' is present on machine %s\n", tagName, systemID)
	})

	t.Run("unassign tag from machines", func(t *testing.T) {
		// Create a test tag
		tagName := "test-assign-unassign-tag"
		err := c.Tags().Create(ctx, tagName)
		assert.Nil(t, err, "Failed to create tag")

		// TODO: Replace with actual machine system ID you want to test with
		// This should be an unlocked machine that can have tags assigned
		systemID := "REPLACE_WITH_MACHINE_SYSTEM_ID_1"

		// Skip test if placeholder values are still present
		if systemID == "REPLACE_WITH_MACHINE_SYSTEM_ID_1" {
			t.Skip("Please replace placeholder machine system ID with actual unlocked machine")
			return
		}

		// Now unassign the tag
		err = c.Tags().Unassign(ctx, tagName, systemID)
		assert.Nil(t, err, "Failed to unassign tag")
		fmt.Printf("Successfully unassigned tag '%s' from machine: %v\n", tagName, systemID)

		// Wait a bit for the unassignment to propagate
		time.Sleep(2 * time.Second)

		// Verify the tag was removed by checking machine details
		machine := c.Machines().Machine(systemID)
		detailedMachine, err := machine.Get(ctx)
		assert.Nil(t, err, "Failed to get machine details for %s", systemID)

		machinetags := detailedMachine.Tags()
		assert.NotContains(t, machinetags, tagName,
			"Tag '%s' should not be present on machine %s after unassign. Machine tags: %v", tagName, systemID, machinetags)
		fmt.Printf("✅ Verified tag '%s' is removed from machine %s\n", tagName, systemID)
	})
}

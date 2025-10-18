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
	"os"
	"testing"

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
		tagName := "test-assign-tag"
		err := c.Tags().Create(ctx, tagName)
		assert.Nil(t, err)

		// Get a list of machines to test with
		machines, err := c.Machines().List(ctx, nil)
		if err != nil || len(machines) == 0 {
			t.Skip("No machines available for testing Assign")
			return
		}

		// Get system IDs of the first few machines (max 2 for testing)
		var systemIDs []string
		for i, machine := range machines {
			if i >= 2 {
				break
			}
			systemIDs = append(systemIDs, machine.SystemID())
		}

		// Assign the tag to the machines
		err = c.Tags().Assign(ctx, tagName, systemIDs)
		assert.Nil(t, err)
		fmt.Printf("Successfully assigned tag '%s' to machines: %v\n", tagName, systemIDs)

		// Verify by getting machine details
		for _, systemID := range systemIDs {
			machine := c.Machines().Machine(systemID)
			detailedMachine, err := machine.Get(ctx)
			if err == nil {
				tags := detailedMachine.Tags()
				fmt.Printf("Machine %s tags: %v\n", systemID, tags)
			}
		}
	})

	t.Run("unassign tag from machines", func(t *testing.T) {
		// Create a test tag
		tagName := "test-unassign-tag"
		err := c.Tags().Create(ctx, tagName)
		assert.Nil(t, err)

		// Get a list of machines to test with
		machines, err := c.Machines().List(ctx, nil)
		if err != nil || len(machines) == 0 {
			t.Skip("No machines available for testing Unassign")
			return
		}

		// Get system IDs of the first few machines (max 2 for testing)
		var systemIDs []string
		for i, machine := range machines {
			if i >= 2 {
				break
			}
			systemIDs = append(systemIDs, machine.SystemID())
		}

		// First assign the tag
		err = c.Tags().Assign(ctx, tagName, systemIDs)
		assert.Nil(t, err)
		fmt.Printf("Assigned tag '%s' to machines: %v\n", tagName, systemIDs)

		// Now unassign the tag
		err = c.Tags().Unassign(ctx, tagName, systemIDs)
		assert.Nil(t, err)
		fmt.Printf("Successfully unassigned tag '%s' from machines: %v\n", tagName, systemIDs)

		// Verify by getting machine details
		for _, systemID := range systemIDs {
			machine := c.Machines().Machine(systemID)
			detailedMachine, err := machine.Get(ctx)
			if err == nil {
				tags := detailedMachine.Tags()
				fmt.Printf("Machine %s tags after unassign: %v\n", systemID, tags)
			}
		}
	})
}

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
	"net/http"
	"net/http/httptest"
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

	t.Run("assign tag to multiple machines", func(t *testing.T) {
		// First, create a test tag
		tagName := "test-assign-multiple-tag"
		err := c.Tags().Create(ctx, tagName)
		assert.Nil(t, err, "Failed to create tag")

		// TODO: Replace with actual machine system IDs you want to test with
		// These should be unlocked machines that can have tags assigned
		systemIDs := []string{"REPLACE_WITH_MACHINE_SYSTEM_ID_1", "REPLACE_WITH_MACHINE_SYSTEM_ID_2"}

		// Skip test if placeholder values are still present
		if systemIDs[0] == "REPLACE_WITH_MACHINE_SYSTEM_ID_1" {
			t.Skip("Please replace placeholder machine system IDs with actual unlocked machines")
			return
		}

		// Assign the tag to all machines in a single request
		err = c.Tags().AssignToMachines(ctx, tagName, systemIDs)
		assert.Nil(t, err, "Failed to assign tag to machines")
		fmt.Printf("Successfully assigned tag '%s' to machines: %v\n", tagName, systemIDs)

		// Wait a bit for the assignment to propagate
		time.Sleep(2 * time.Second)

		for _, systemID := range systemIDs {
			machine := c.Machines().Machine(systemID)
			detailedMachine, err := machine.Get(ctx)
			assert.Nil(t, err, "Failed to get machine details for %s", systemID)
			tags := detailedMachine.Tags()
			assert.Contains(t, tags, tagName,
				"Tag '%s' not found on machine %s. Machine tags: %v", tagName, systemID, tags)
			fmt.Printf("✅ Verified tag '%s' is present on machine %s\n", tagName, systemID)
		}
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

func TestTagsAssignToMachines(t *testing.T) {
	ctx := context.Background()

	t.Run("posts one add value per system ID in a single request", func(t *testing.T) {
		var gotMethod, gotPath string
		var gotAdd []string
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			gotMethod = r.Method
			gotPath = r.URL.Path
			assert.Nil(t, r.ParseForm())
			gotAdd = r.PostForm["add"]
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{}")
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		err := c.Tags().AssignToMachines(ctx, "gpu-tag", []string{"abc123", "def456", "ghi789"})
		assert.Nil(t, err)
		assert.Equal(t, 1, requestCount)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/api/2.0/tags/gpu-tag/op-update_nodes", gotPath)
		assert.Equal(t, []string{"abc123", "def456", "ghi789"}, gotAdd)
	})

	t.Run("skips empty system IDs", func(t *testing.T) {
		var gotAdd []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Nil(t, r.ParseForm())
			gotAdd = r.PostForm["add"]
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{}")
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		err := c.Tags().AssignToMachines(ctx, "gpu-tag", []string{"", "abc123", ""})
		assert.Nil(t, err)
		assert.Equal(t, []string{"abc123"}, gotAdd)
	})

	t.Run("no-op without a request when tag name is empty", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		err := c.Tags().AssignToMachines(ctx, "", []string{"abc123"})
		assert.Nil(t, err)
		assert.Equal(t, 0, requestCount)
	})

	t.Run("no-op without a request when no usable system IDs", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		err := c.Tags().AssignToMachines(ctx, "gpu-tag", nil)
		assert.Nil(t, err)

		err = c.Tags().AssignToMachines(ctx, "gpu-tag", []string{"", ""})
		assert.Nil(t, err)

		assert.Equal(t, 0, requestCount)
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "No Tag matches the given query.")
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		err := c.Tags().AssignToMachines(ctx, "missing-tag", []string{"abc123"})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "status: 404")
	})

	t.Run("escapes tag name in the path", func(t *testing.T) {
		var gotEscapedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotEscapedPath = r.URL.EscapedPath()
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{}")
		}))
		defer server.Close()

		c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

		err := c.Tags().AssignToMachines(ctx, "gpu tag", []string{"abc123"})
		assert.Nil(t, err)
		assert.Equal(t, "/api/2.0/tags/gpu%20tag/op-update_nodes", gotEscapedPath)
	})
}

// Regression test: Tag.UnmarshalJSON used to populate comment from the
// kernel_opts field, so Comment() returned kernel options instead of the
// tag's comment.
func TestTagsList_UnmarshalsComment(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[
		  {
		    "name": "gpu",
		    "definition": "//node[@class=\"display\"]",
		    "comment": "machines with GPUs",
		    "kernel_opts": "console=tty0",
		    "resource_uri": "/MAAS/api/2.0/tags/gpu/"
		  }
		]`)
	}))
	defer server.Close()

	c := NewAuthenticatedClientSet(server.URL, "consumer:token:secret")

	tags, err := c.Tags().List(ctx)
	assert.Nil(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, "gpu", tags[0].Name())
	assert.Equal(t, "machines with GPUs", tags[0].Comment())
	assert.Equal(t, "console=tty0", tags[0].KernelOpts())
	assert.Equal(t, `//node[@class="display"]`, tags[0].Definition())
}

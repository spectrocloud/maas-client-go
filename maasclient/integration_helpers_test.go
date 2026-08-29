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
	"time"
)

// requireMAASIntegration skips the test when MAAS_ENDPOINT or MAAS_API_KEY is unset.
func requireMAASIntegration(t *testing.T) ClientSetInterface {
	t.Helper()

	endPoint := os.Getenv("MAAS_ENDPOINT")
	apiKey := os.Getenv("MAAS_API_KEY")
	if endPoint == "" || apiKey == "" {
		t.Skip("MAAS_ENDPOINT and MAAS_API_KEY must be set for integration tests")
	}

	return NewAuthenticatedClientSet(endPoint, apiKey)
}

// testDNSDomain returns the DNS domain to use in integration tests.
// It prefers MAAS_TEST_DNS_DOMAIN; otherwise it uses the first domain from MAAS.
func testDNSDomain(t *testing.T, c ClientSetInterface, ctx context.Context) string {
	t.Helper()

	if domain := os.Getenv("MAAS_TEST_DNS_DOMAIN"); domain != "" {
		return domain
	}

	domains, err := c.Domains().List(ctx)
	if err != nil {
		t.Fatalf("list domains: %v", err)
	}
	for _, d := range domains {
		if name := d.Name(); name != "" {
			return name
		}
	}

	t.Skip("no DNS domain found in MAAS; set MAAS_TEST_DNS_DOMAIN")
	return ""
}

// discoverDNSResourceFQDN returns an existing DNS resource FQDN for read-only tests.
// It prefers MAAS_TEST_DNS_FQDN; otherwise it uses the first listed resource.
func discoverDNSResourceFQDN(t *testing.T, c ClientSetInterface, ctx context.Context) string {
	t.Helper()

	if fqdn := os.Getenv("MAAS_TEST_DNS_FQDN"); fqdn != "" {
		return fqdn
	}

	resources, err := c.DNSResources().List(ctx, nil)
	if err != nil {
		t.Fatalf("list dns resources: %v", err)
	}
	for _, r := range resources {
		if fqdn := r.FQDN(); fqdn != "" {
			return fqdn
		}
	}

	t.Skip("no DNS resources found; set MAAS_TEST_DNS_FQDN")
	return ""
}

func uniqueDNSResourceName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func dnsFQDN(name, domain string) string {
	return name + "." + domain
}

// discoverMachineSystemID returns a machine system ID for read-only integration tests.
// It prefers MAAS_TEST_MACHINE_SYSTEM_ID; otherwise it uses the first listed machine.
func discoverMachineSystemID(t *testing.T, c ClientSetInterface, ctx context.Context) string {
	t.Helper()

	if systemID := os.Getenv("MAAS_TEST_MACHINE_SYSTEM_ID"); systemID != "" {
		return systemID
	}

	machines, err := c.Machines().List(ctx, nil)
	if err != nil {
		t.Fatalf("list machines: %v", err)
	}
	for _, m := range machines {
		if systemID := m.SystemID(); systemID != "" {
			return systemID
		}
	}

	t.Skip("no machines found; set MAAS_TEST_MACHINE_SYSTEM_ID")
	return ""
}

// discoverTestZone returns a zone name for integration tests.
// It prefers MAAS_TEST_MACHINE_ZONE; otherwise it uses the first zone from MAAS.
func discoverTestZone(t *testing.T, c ClientSetInterface, ctx context.Context) string {
	t.Helper()

	if zone := os.Getenv("MAAS_TEST_MACHINE_ZONE"); zone != "" {
		return zone
	}

	zones, err := c.Zones().List(ctx)
	if err != nil {
		t.Fatalf("list zones: %v", err)
	}
	for _, z := range zones {
		if name := z.Name(); name != "" {
			return name
		}
	}

	t.Skip("no zones found; set MAAS_TEST_MACHINE_ZONE")
	return ""
}

// requireMachineMutations skips tests that allocate, deploy, or update machines.
func requireMachineMutations(t *testing.T) {
	t.Helper()

	if os.Getenv("MAAS_TEST_ALLOW_MACHINE_MUTATIONS") != "1" {
		t.Skip("machine mutation tests disabled; set MAAS_TEST_ALLOW_MACHINE_MUTATIONS=1 to enable")
	}
}

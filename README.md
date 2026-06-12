# maas-client-go

MAAS client for Go.

## Development

### Prerequisites

- Go 1.24
- Make
- Golangci-lint

### Setup

```bash
make env
```
## Testing

Set `MAAS_ENDPOINT` and `MAAS_API_KEY` in your environment.

```bash
export MAAS_ENDPOINT="http://10.0.0.1:5240/MAAS"
export MAAS_API_KEY="abc123:def456:ghi789"
```

`make test` runs the full test suite. `MAAS_ENDPOINT` and `MAAS_API_KEY` must be set.

Optional variables for integration tests that need environment-specific data:

| Variable | Used by | Description |
|----------|---------|-------------|
| `MAAS_TEST_DNS_DOMAIN` | DNS tests | DNS domain for create/delete tests (e.g. `maas`). Defaults to the first domain returned by MAAS. |
| `MAAS_TEST_DNS_FQDN` | DNS tests | Existing DNS record to use for FQDN filter tests. Defaults to the first listed resource. |
| `MAAS_TEST_BOOT_RESOURCE_FILE` | Boot resource import test | Local path to a boot resource archive. Test is skipped when unset or the file does not exist. |
| `MAAS_TEST_BOOT_RESOURCE_ID` | Boot resource get-by-id test | Boot resource ID to fetch. Defaults to the first resource from list. |
| `MAAS_TEST_BOOT_RESOURCE_NAME` | Boot resource import test | Name for the imported resource. Defaults to a unique generated name. |
| `MAAS_TEST_BOOT_RESOURCE_ARCH` | Boot resource import test | Architecture for the import. Defaults to `amd64/generic`. |
| `MAAS_TEST_MACHINE_SYSTEM_ID` | Machine tests | Machine to use. Defaults to the first machine from list. |
| `MAAS_TEST_MACHINE_ZONE` | Machine allocate test | Zone for allocation. Defaults to the first zone from list. |
| `MAAS_TEST_ALLOW_MACHINE_MUTATIONS` | Machine allocate/deploy/update tests | Set to `1` to enable tests that allocate, deploy, or modify machines. |
| `MAAS_TEST_DISTRO_SERIES` | Machine deploy test | Distro series to deploy. Required when machine mutations are enabled. |

Some integration tests also assume specific data in your MAAS environment (hardcoded machine system IDs, zones, boot resources, and so on). Those may fail even with valid credentials unless your cluster matches the test fixtures.

A small set of tests use `httptest` and run offline with no MAAS instance:

```bash
go test ./maasclient -run 'TestAuthHeader|TestWithTags|TestTagsAssignToMachines|TestNodeDevicesList'
```

## Usage

```
	c := NewAuthenticatedClientSet(os.Getenv("MAAS_ENDPOINT"), os.Getenv("MAAS_API_KEY"))

	ctx := context.Background()

	// List DNS Resources
	res, err := c.DNSResources().List(ctx, nil)



	// List DNS Resources filtered by fqdn
	filters := ParamsBuilder().Add(FQDNKey, "bad-doesntexist.maas")
	res, err := c.DNSResources().List(ctx, filters)



	// Create DNS Resource
	res, err := c.DNSResources().
		Builder().
		WithFQDN("test-unit1.maas.sc").
		WithAddressTTL("10").Create(ctx)


	// Update DNS Resource
	err = res.Modifier().
		SetIPAddresses([]string{"1.2.3.4", "5.6.7.8"}).
		Modify(ctx)


	// Get DNS Resource by ID
	res2 := c.DNSResources().DNSResource(res.ID())


	// Delete DNS Resource
	err = res.Delete(ctx)

```

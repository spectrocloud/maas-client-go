# maas-client-go
MAAS client for GO

Usage

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

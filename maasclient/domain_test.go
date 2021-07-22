package maasclient_test

import (
	"github.com/spectrocloud/maas-client-go/maasclient"
	"os"
	"testing"
)

func TestMaasResourceDomain(t *testing.T) {

	client := maasclient.NewClient(os.Getenv(MAAS_ENDPOINT), os.Getenv(MAAS_APIKEY))
	pools, err := client.GetDomain()
	if err != nil {
		t.Error(err.Error())
	}
	for _, pool := range pools {
		if pool.Name == "" {
			t.Error("Domain name is empty")
		}
	}
	t.Logf("Domain %#v", err)

}
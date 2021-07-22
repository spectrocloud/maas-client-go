package maasclient_test

import (
	"github.com/spectrocloud/maas-client-go/maasclient"
	"os"
	"testing"
)

func TestMaasResourceResourcePool(t *testing.T) {

	client := maasclient.NewClient(os.Getenv(MAAS_ENDPOINT), os.Getenv(MAAS_APIKEY))
	pools, err := client.GetPools()
	if err != nil {
		t.Error(err.Error())
	}
	for _, pool := range pools {
		if pool.Name == "" {
			t.Error("Resource pool name is empty")
		}
	}
	t.Logf("Resource pool %#v", err)

}
package maasclient_test

import (
	"github.com/spectrocloud/maas-client-go/maasclient"
	"os"
	"testing"
)

func TestMaasResourceSubnet(t *testing.T) {

	client := maasclient.NewClient(os.Getenv(MAAS_ENDPOINT), os.Getenv(MAAS_APIKEY))
	subnets, err := client.GetSubnets()
	if err != nil {
		t.Error(err.Error())
	}
	for _, subnet := range subnets {
		if subnet.Name == "" {
			t.Error("Subnet name is empty")
		}
	}
	t.Logf("Subnet %#v", err)

}
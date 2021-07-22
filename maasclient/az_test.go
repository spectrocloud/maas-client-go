package maasclient_test

import (
	"github.com/spectrocloud/maas-client-go/maasclient"
	"os"
	"testing"
)

func TestMaasResourceAz(t *testing.T) {

	client := maasclient.NewClient(os.Getenv(MAAS_ENDPOINT), os.Getenv(MAAS_APIKEY))
	azs, err := client.GetZones()
	if err != nil {
		t.Error(err.Error())
	}
	for _, az := range azs {
		if az.Name == "" {
			t.Error("Az name is empty")
		}
	}
	t.Logf("Azs %#v", err)

}
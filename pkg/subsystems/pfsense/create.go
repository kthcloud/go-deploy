package pfsense

import (
	"context"
	"fmt"
	"github.com/saffronjam/pfsense-client"
	"go-deploy/utils/requestutils"
	"net/http"
)

func OpenPort() {
	addBasicAuth := func(client *pfsense.Client) error {
		client.RequestEditors = append(client.RequestEditors, func(ctx context.Context, req *http.Request) error {
			req.SetBasicAuth("admin", "nerexasa")
			return nil
		})
		return nil
	}

	client, err := pfsense.NewClient("https://fw.kthcloud.com", addBasicAuth)

	if err != nil {
		fmt.Println(err)
		return
	}

	apply := true
	disabled := false
	srcdst := "8888"
	localport := "22"
	desc := "description"

	nordr := false

	reqBody := pfsense.APIFirewallNATOutboundPortForwardCreateJSONRequestBody{
		Apply:         &apply,
		Descr:         &desc,
		Disabled:      &disabled,
		Dst:           "130.237.83.248",
		Dstport:       &srcdst,
		Interface:     "WAN",
		LocalPort:     &localport,
		Natreflection: nil,
		Nordr:         &nordr,
		Nosync:        &nordr,
		Protocol:      "tcp",
		Src:           "",
		Srcport:       nil,
		Target:        "172.31.1.69",
		Top:           nil,
	}

	create, err := client.APIFirewallNATOutboundPortForwardCreate(context.TODO(), reqBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	body, _ := requestutils.ReadBody(create.Body)
	fmt.Println(string(body))

	fmt.Println(create)

}

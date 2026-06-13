// CA机构搜索示例
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/crt.sh-skills/pkg/crtsh"
	"log"
)

func main() {
	client := crtsh.NewClient()

	// 搜索指定CA签发的证书
	params := crtsh.QueryParams{
		SearchType: "CAID",
		CAID:       "12345",
		PageSize:   50,
	}

	certs, _, err := client.SearchCertificates(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("CA ID 12345 签发了 %d 个证书:\n", len(certs))
	for _, cert := range certs {
		fmt.Printf("- %s (有效期至 %s)\n",
			cert.Domains[0],
			cert.NotAfter.Format("2006年01月"),
		)
	}
}

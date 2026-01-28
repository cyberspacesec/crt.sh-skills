// 分页处理示例
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/go-crt.sh/pkg/crtsh"
	"log"
)

func main() {
	client := crtsh.NewClient()
	params := crtsh.QueryParams{
		Q:          "bank",
		SearchType: "CN",
		PageSize:   20,
	}

	// 收集所有结果
	var allCerts []crtsh.Certificate

	for {
		certs, pagination, err := client.SearchCertificates(context.Background(), params)
		if err != nil {
			log.Fatal(err)
		}

		allCerts = append(allCerts, certs...)
		fmt.Printf("已获取 %d 条结果...\n", len(allCerts))

		if pagination.NextPage == 0 {
			break
		}
		params.Page = pagination.NextPage
	}

	fmt.Printf("\n最终获得 %d 个含 'bank' 的证书\n", len(allCerts))
}

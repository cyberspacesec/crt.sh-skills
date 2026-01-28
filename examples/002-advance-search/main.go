// 高级搜索示例
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/go-crt.sh/pkg/crtsh"
	"log"
	"time"
)

func main() {
	client := crtsh.NewClient()
	client.Debug = true // 启用调试模式

	// 使用精确匹配搜索
	params := crtsh.QueryParams{
		SearchType:     "sha256",
		SHA256:         "a1b2c3...",
		ExcludeExpired: true,
		Linter:         "zlint",
		LintType:       "issues",
	}

	certs, _, err := client.SearchCertificates(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	if len(certs) == 0 {
		fmt.Println("未找到匹配的证书")
		return
	}

	cert := certs[0]
	fmt.Printf("[详细报告]\n颁发者ID: %d\n提交时间: %s\n",
		cert.IssuerCAID,
		cert.EntryTimestamp.Format(time.RFC3339),
	)
}

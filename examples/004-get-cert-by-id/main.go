// 按ID获取证书示例
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/crt.sh-skills/pkg/crtsh"
	"log"
)

func main() {
	client := crtsh.NewClient()

	cert, err := client.GetCertificateByID(context.Background(), 41794150)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("证书 #%d 详细信息:\n", cert.ID)
	fmt.Println("原始名称字段:", cert.RawNameValue)
	fmt.Println("解析后的域名:")
	for _, domain := range cert.Domains {
		fmt.Println("  ", domain)
	}
	fmt.Printf("完整元数据:\n%+v\n", cert)
}

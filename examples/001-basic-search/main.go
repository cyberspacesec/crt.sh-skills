// 基础搜索示例
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

	// 设置搜索参数
	params := crtsh.QueryParams{
		Q:           "example.com",
		SearchType:  "c", // 证书搜索
		Page:        1,
		PageSize:    10,
		Deduplicate: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 执行搜索
	certs, pagination, err := client.SearchCertificates(ctx, params)
	if err != nil {
		log.Fatalf("搜索失败: %v", err)
	}

	fmt.Printf("找到 %d 个证书 (第%d页，共%d页)\n",
		len(certs),
		pagination.CurrentPage,
		pagination.CurrentPage+(pagination.NextPage-pagination.CurrentPage),
	)

	for _, cert := range certs {
		fmt.Printf("证书ID: %d\n序列号: %s\n有效期: %s 至 %s\n域名列表: %v\n\n",
			cert.ID,
			cert.SerialNumber,
			cert.NotBefore.Format("2006-01-02"),
			cert.NotAfter.Format("2006-01-02"),
			cert.Domains,
		)
	}

	// Output:
	// 找到 33 个证书 (第1页，共0页)
	//证书ID: 16593729128
	//序列号: 00d389b7d7936a9a5efbd697c8af3ecbf9
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com]
	//
	//证书ID: 16593729744
	//序列号: 00d389b7d7936a9a5efbd697c8af3ecbf9
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com]
	//
	//证书ID: 16417331563
	//序列号: 0ad893bafa68b0b7fb7a404f06ecaf9a
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com]
	//
	//证书ID: 16488405764
	//序列号: 00f89cde7fe576b4ebf16bfd7a332f75ae
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 16488405627
	//序列号: 00f89cde7fe576b4ebf16bfd7a332f75ae
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 16233306772
	//序列号: 0202d62a240d61737b5dbf132c87fcf3
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 16231429270
	//序列号: 0ad893bafa68b0b7fb7a404f06ecaf9a
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com]
	//
	//证书ID: 16228519060
	//序列号: 0202d62a240d61737b5dbf132c87fcf3
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 12337892544
	//序列号: 075bcef30689c8addf13e51af4afe187
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 11920382870
	//序列号: 075bcef30689c8addf13e51af4afe187
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 8913351873
	//序列号: 0c1fcb184518c7e3866741236d6b73f1
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 8396709327
	//序列号: 0c1fcb184518c7e3866741236d6b73f1
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 6359075900
	//序列号: 0faa63109307bc3d414892640ccd4d9a
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 6342480680
	//序列号: 0faa63109307bc3d414892640ccd4d9a
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 5813209289
	//序列号: 025216e1c4998e2632aa5d1da985b43c
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 5771467708
	//序列号: 025216e1c4998e2632aa5d1da985b43c
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 3704614715
	//序列号: 0fbe08b0854d05738ab0cce1c9afeec9
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 3692510597
	//序列号: 0fbe08b0854d05738ab0cce1c9afeec9
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 2854376823
	//序列号: 0fd078dd48f1a2bd4d0f2ba96b6038fe
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 2854374595
	//序列号: 0fd078dd48f1a2bd4d0f2ba96b6038fe
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 2854374664
	//序列号: 0fd078dd48f1a2bd4d0f2ba96b6038fe
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 987119772
	//序列号: 0fd078dd48f1a2bd4d0f2ba96b6038fe
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 984858191
	//序列号: 0fd078dd48f1a2bd4d0f2ba96b6038fe
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 34083306
	//序列号: 06fb0a7d9e401925b755cf0b76da3585
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [subjectname@example.com]
	//
	//证书ID: 34001389
	//序列号: 471d7f510e46e7f5304feb7a7636be0d
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [subjectname@example.com]
	//
	//证书ID: 24564717
	//序列号: 6c2dae34a020836b1f86eeb7d1e52f51
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com m.testexample.com www.example.com]
	//
	//证书ID: 24560643
	//序列号: 1286c6a95e41d9687326dd68ee421416
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com]
	//
	//证书ID: 24560621
	//序列号: 75547e6f9d1c6f1b60227e84c9d83203
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com m.example.com www.example.com]
	//
	//证书ID: 24558997
	//序列号: 3553dc6830920c98dec99ec9629fdb93
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [dev.example.com example.com products.example.com support.example.com www.example.com]
	//
	//证书ID: 10557607
	//序列号: 0e64c5fbc236ade14b172aeb41c78cb0
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 5857507
	//序列号: 0411de8f53b462f6a5a861b712ec6b59
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com www.example.com]
	//
	//证书ID: 8506962125
	//序列号: 1ac1e693c87d36563a92ca145c87bbc26fd49f4c
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [example.com user@example.com]
	//
	//证书ID: 10570508844
	//序列号: 1000
	//有效期: 0001-01-01 至 0001-01-01
	//域名列表: [AS207960 Test Intermediate - example.com]

}

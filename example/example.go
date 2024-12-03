package main

import (
	"fmt"

	"github.com/alexzanda/vyos-client"
)

func main() {
	client := vyos.NewClient(vyos.Config{Host: "http://192.170.2.249:8080", SkipTLS: true})
	// 设置地址
	//fmt.Println(client.SetAddress("eth1", "10.10.7.163/16"))

	// 删除地址
	//fmt.Println(client.DeleteAddress("eth1", "10.10.7.163/16"))

	// 添加snat
	fmt.Println(client.AddSnat(10, "10.20.122.10", "8.8.8.8", "eth2"))

	// 删除snat
	//fmt.Println(client.DeleteSnat(10))

	// 添加dnat
	fmt.Println(client.AddDnat(10, "8.8.8.8", "10.20.122.111", "eth2"))

	// 删除dnat
	//fmt.Println(client.DeleteDnat(10))

	// 添加路由
	//fmt.Println(client.AddRoute("10.30.122.10/24", "1::122:1", 6))

	//// 展示所有配置
	//fmt.Println(client.ShowConfiguration())
	//
	//// 持久保存
	fmt.Println(client.SaveConfig())
}

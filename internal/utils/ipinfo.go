package utils

import (
	"log"
	"net"
	"strings"

	"tfile/internal/config"
)

func ShowIPinfo() {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, interf := range interfaces {
		if interf.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := interf.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			// 解析 IP 地址并排除子网掩码长度后缀
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				log.Println(interf.Name, "| Link: http://"+addr.String()+":"+config.ServerPort)
				continue
			}
			// 排除 fe80 开头的 IPv6 地址
			if ip.To4() == nil && strings.HasPrefix(ip.String(), "fe80") {
				continue
			}
			log.Println(interf.Name, "| Link: http://"+ip.String()+":"+config.ServerPort)
		}
	}
}
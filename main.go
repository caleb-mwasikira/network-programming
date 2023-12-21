package main

import (
	"fmt"
	"log"

	"bellweathertech.com/network-programming/ip"
)

func main() {
	ipv6Str := "2001:0db8:3333:4444:5555:6666:7777:8888"
	ipv6Binary, err := ip.ConvertIpv6ToBinary(ipv6Str)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("IPv6 string [%v]\n", ipv6Str)
	fmt.Printf("IPv6 binary [%v]\n", ipv6Binary)
	return
}

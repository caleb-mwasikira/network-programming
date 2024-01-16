package main

import (
	"fmt"
	"log"

	"bellweathertech.com/network-programming/ip"
)

func main() {
	ipv6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	ipv6Binary, err := ip.IPv6StringToBinary(ipv6)
	if err != nil {
		log.Fatal(err)
	}

	shortenedIpv6 := ip.ShortenIPv6Address(ipv6)

	fmt.Printf("IPv6 string %v\n", ipv6)
	fmt.Printf("Shortened IPv6 string %v\n", shortenedIpv6)
	fmt.Printf("IPv6 to binary conversion %v\n", ipv6Binary)
}

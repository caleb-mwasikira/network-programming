package ip

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

/*
Splits an IPv6 address into its 8 colon-separated
groups of 16bit segments called hextets

For example, the IPv6 address 2001:0db8:3333:4444:5555:6666:7777:8888

Will be converted to:

	[]string{
		'2001', '0db8', '3333', '4444', '5555', '6666', '7777',
		'8888'
	}
*/
func splitIPv6IntoSegments(ipv6 string) ([]string, error) {
	numSegments := 8
	ipv6 = strings.ReplaceAll(strings.Trim(ipv6, " "), ":", " ")
	segments := strings.Fields(ipv6)

	if len(segments) != numSegments {
		return nil, fmt.Errorf("Expected %v segments from ipv6 address, but got %v instead\n", numSegments, len(segments))
	}

	return segments, nil
}

/*
Checks if an address is a valid IPv6 address
*/
func isValidIPv6(ipv6 string) (bool, error) {
	segments, err := splitIPv6IntoSegments(ipv6)
	if err != nil {
		return false, err
	}

	// Check if each segment is a valid hexadecimal value
	for _, segment := range segments {
		hexPattern := regexp.MustCompile("^[0-9a-fA-F]+$")
		if !hexPattern.MatchString(segment) {
			return false, fmt.Errorf("Segment %v of ipv6 address is not a valid hexadecimal value\n", segment)
		}
	}

	return true, nil
}

func ConvertIpv6ToBinary(ipv6 string) (string, error) {
	if ok, err := isValidIPv6(ipv6); !ok {
		return "", err
	}

	segments, err := splitIPv6IntoSegments(ipv6)
	if err != nil {
		return "", err
	}

	ipv6BinaryStr := ""

	for _, segment := range segments {
		// Convert each segment to binary and append to binary string
		hexInt, ok := new(big.Int).SetString(segment, 16)
		if !ok {
			return "", fmt.Errorf("Invalid hexadecimal string: %s\n", segment)
		}

		ipv6BinaryStr += fmt.Sprintf("%016b", hexInt) + ":"
	}

	return ipv6BinaryStr, nil
}

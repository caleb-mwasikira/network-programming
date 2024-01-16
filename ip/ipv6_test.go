package ip

import (
	"fmt"
	"testing"
)

func TestIPv6StringToBinary(t *testing.T) {
	tests := []struct {
		address    string
		expected   string
		shouldFail bool
	}{
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "0010000000000001:0000110110111000:1000010110100011:0000000000000000:0000000000000000:1000101000101110:0000001101110000:0111001100110100", false}, // Valid IPv6 address
		{"::1", "0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000001", false},                                     // Valid IPv6 loopback address
		{"fe80::1%eth0", "1111111010000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000001", false},                            // Valid IPv6 address with zone identifier
		{"invalid", "", true},     // Invalid IPv6 address
		{"192.168.0.1", "", true}, // Not an IPv6 address
	}

	for _, test := range tests {
		t.Run(test.address, func(t *testing.T) {
			got, err := IPv6StringToBinary(test.address)

			if test.shouldFail {
				if err == nil {
					t.Errorf("expected an error for invalid address %v", test.address)
				}

			} else {
				if err != nil {
					t.Errorf("unexpected error %v while converting valid address %v", err, test.address)
				}

				if got != test.expected {
					t.Errorf("expected binary ipv6 address %v but got %v instead", test.expected, string(got))
				}
			}
		})
	}
}

func TestShortenIPv6Address(t *testing.T) {
	tests := []struct {
		address  string
		expected string
	}{
		{"2001:0db8:85a3:0000:0001:8a2e:0370:7334", "2001:db8:85a3:0:1:8a2e:370:7334"}, // Should remove leading zeros
		{"2001:0000:0000:0001:0000:0000:0000:0000", "2001:0:0:1::"},                    // Multiple consecutive zero groups of different sizes; should compress largest consecutive zero group
		{"fe80::1", "fe80::1"},
		{"3ffe:1900:4545:3:200:f8ff:fe21:67cf", "3ffe:1900:4545:3:200:f8ff:fe21:67cf"},
		{"::", "::"},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("Shortening IPv6 address %v", test.address)

		t.Run(testName, func(t *testing.T) {
			got := ShortenIPv6Address(test.address)

			if got != test.expected {
				t.Errorf("expected shortened IPv6 address %v but got %v instead\n", test.expected, got)
			}
		})
	}
}

func TestExpandIPv6Address(t *testing.T) {
	tests := []struct {
		address  string
		expected string
	}{
		{"2001:db8:85a3::8a2e:370:7334", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		{"fe80::1", "fe80:0000:0000:0000:0000:0000:0000:0001"},
		{"3ffe:1900:4545:3:200:f8ff:fe21:67cf", "3ffe:1900:4545:0003:0200:f8ff:fe21:67cf"},
		{"fd12:3456:789a:1::1", "fd12:3456:789a:0001:0000:0000:0000:0001"},
		{"::", "0000:0000:0000:0000:0000:0000:0000:0000"},
		{"2001:0:0:1::", "2001:0000:0000:0001:0000:0000:0000:0000"},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("Expanding IPv6 address %v", test.address)

		t.Run(testName, func(t *testing.T) {
			got := ExpandIPv6Address(test.address)

			if got != test.expected {
				t.Errorf("expected expanded IPv6 address %v but got %v instead\n", test.expected, got)
			}
		})
	}
}

func TestStripLeadingZerosFromSegment(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{"2001", "2001"},
		{"0db8", "db8"},
		{"0005", "5"},
		{"0000", "0"},
		{"0370", "370"},
		{"0600", "600"},
		{"ef01", "ef01"},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("Stripping leading zeros from str %v", test.str)

		t.Run(testName, func(t *testing.T) {
			got := stripLeadingZerosFromSegment(test.str)

			if got != test.expected {
				t.Errorf("expected %v but got %v instead", test.expected, got)
			}
		})
	}
}

func TestIndexOfIPv6ZeroGroup(t *testing.T) {
	tests := []struct {
		address    string
		startIndex int
		stopIndex  int
	}{
		{"2001:0db8:0000:0042:0000:8a2e:0370:7334", -1, -1},
		{"fe80:0000:0000:0000:0000:0000:0000:0001", 1, 6},
		{"3ffe:1900:0000:0003:0200:f8ff:fe21:67cf", -1, -1},
		{"2001:db8:0000:0000:0000:0000:0000:0001", 2, 6},
		{"fd12:3456:789a:0001:0000:0000:0000:0001", 4, 6},
		{"abcd:ef01:2345:6789:abcd:ef01:2345:6789", -1, -1},
		{"0000:0000:0000:0000:0000:0000:0000:0000", 0, 7},
		{"2001:0000:0000:0001:0000:0000:0000:0000", 4, 7},
		{"1234:0000:5678:0000:abcd:0000:ef01:0000", -1, -1},
		{"::1", -1, -1},
		{"2001:0:0:1::", 1, 2},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("Grouping zeros in IPv6 address %v", test.address)

		t.Run(testName, func(t *testing.T) {
			gotStartIndex, gotStopIndex := indexOfIPv6ZeroGroup(test.address)

			if gotStartIndex != int(test.startIndex) || gotStopIndex != int(test.stopIndex) {
				t.Errorf("expected grouping start and stop indexes of (%v, %v) but got (%v, %v) instead", test.startIndex, test.stopIndex, gotStartIndex, gotStopIndex)
			}
		})
	}
}

package ip

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

func IPv6StringToBinary(ipv6 string) (string, error) {
	ipv6, err := validateIPv6Address(ipv6)
	if err != nil {
		return "", err
	}

	var ipv6Binary []string
	segments := strings.Split(ipv6, ":")
	for _, segment := range segments {
		// convert hexadecimal to binary/bytes
		hexInt, ok := new(big.Int).SetString(segment, 16)
		if !ok {
			return "", fmt.Errorf("invalid hexadecimal string %v detected in address %v", segment, ipv6)
		}

		segment = fmt.Sprintf("%016b", hexInt.Int64())
		ipv6Binary = append(ipv6Binary, segment)
	}

	return strings.Join(ipv6Binary, ":"), nil
}

/*
Shortens an IPv6 address.
First, by removing all leading zeros in each hextet.
The address:

	fd00:4700:0010:0000:0000:0000:6814:d103

will be shortened to:

	fd00:4700:10:0:0:0:6814:d103.

Second, replace the largest or leftmost group of consecutive, zero-value
hextets with double colons, producing the shorter

	fd00:4700:10::6814:d103.

If your address has more than one group of consecutive zero-value hextets with
the same size, you can only replace the leftmost zero group
*/
func ShortenIPv6Address(ipv6 string) string {
	var segments []string

	for _, segment := range strings.Split(ipv6, ":") {
		stripped := stripLeadingZerosFromSegment(segment)
		segments = append(segments, stripped)
	}

	ipv6 = strings.Join(segments, ":")
	startIndex, stopIndex := indexOfIPv6ZeroGroup(ipv6)
	if startIndex == -1 || stopIndex == -1 {
		return ipv6
	}

	left := strings.Join(segments[:startIndex], ":")
	right := strings.Join(segments[stopIndex+1:], ":")

	return left + "::" + right
}

func ExpandIPv6Address(ipv6 string) string {
	var segments []string

	left, right, found := strings.Cut(ipv6, "::")
	if found {
		leftSegments := strings.Split(strings.Trim(left, " "), ":")
		rightSegments := strings.Split(right, ":")

		segments = append(segments, leftSegments...)

		// number of zero groups between shortened ipv6 double colons ::
		numSegments := 8
		numZeroGroups := numSegments - (len(leftSegments) + len(rightSegments))

		for i := 0; i < numZeroGroups; i++ {
			segments = append(segments, "0")
		}

		segments = append(segments, rightSegments...)
	} else {
		segments = strings.Split(ipv6, ":")
	}

	// add leading zeros to each segment till 4 digit hexadecimal value
	for i, segment := range segments {
		str := fmt.Sprintf("%04v", segment)
		segments[i] = str
	}
	return strings.Join(segments, ":")
}

func stripLeadingZerosFromSegment(str string) string {
	result := str

	for i := 0; i < len(str); i++ {
		if len(result) == 1 {
			return result
		}

		if result[0] == '0' {
			result = result[i+1:]
		} else {
			return result
		}
	}

	return result
}

/*
Finds the index of the leftmost consecutive zero group existing within an IPv6 address.
Returns -1, -1 if there are no zero groups or if existing
zero groups are not consecutive

For example:

"fd12:3456:789a:0001:0000:0000:0000:0001"
returns 4, 6

"abcd:ef01:2345:6789:abcd:ef01:2345:6789"
returns -1, -1 as there are no zero groups

"2001:0db8:0000:0042:0000:8a2e:0370:7334"
returns -1, -1 as the zero groups at index 2 and 4 are not consecutive; they exist on their own
*/
func indexOfIPv6ZeroGroup(ipv6 string) (int, int) {
	ipv6 = strings.Trim(ipv6, " ")
	segments := strings.Split(ipv6, ":")

	startIndex := -1
	stopIndex := -1
	type zeroGroup struct {
		startIndex int
		stopIndex  int
		length     int
	}
	groups := make([]zeroGroup, 0)

	/*
		Check is segment contains group of zeros i.e 0, 00, 000 or 0000
		Explanation:

		^: Asserts the start of the string.
		0{1,4}: Matches 1 to 4 consecutive zeros at the beginning of the string.
		$: Asserts the end of the string.
	*/
	zeroGroupPattern := regexp.MustCompile("^0{1,4}$")

	for index, segment := range segments {
		isZeroGroup := zeroGroupPattern.MatchString(segment)
		isConsecutive := startIndex != stopIndex

		if isZeroGroup {
			if startIndex == -1 {
				startIndex = index
				stopIndex = index
			} else {
				stopIndex++
			}

			onLastIndex := index == len(segments)-1
			if onLastIndex && isConsecutive {
				groups = append(groups, zeroGroup{
					startIndex: startIndex,
					stopIndex:  stopIndex,
					length:     (stopIndex - startIndex) + 1,
				})
			}

		} else {
			// zero group was already found but now terminating
			// as current segment not zero group
			if startIndex != -1 {
				if isConsecutive {
					groups = append(groups, zeroGroup{
						startIndex: startIndex,
						stopIndex:  stopIndex,
						length:     (stopIndex - startIndex) + 1,
					})
				}

				// reset
				startIndex = -1
				stopIndex = -1
			}

		}
	}

	if len(groups) == 0 {
		return -1, -1
	} else {
		// Find largest or leftmost consecutive zero group
		largestGroup := groups[0]
		for _, group := range groups {
			if group.length > largestGroup.length {
				largestGroup = group
			}
		}
		return largestGroup.startIndex, largestGroup.stopIndex
	}
}

func validateIPv6Address(ipv6 string) (string, error) {
	// strip zone identifier (like %eth0) on IPv6 address, if any
	ipv6, _, _ = strings.Cut(ipv6, "%")

	isShortened := strings.Contains(ipv6, "::")
	if isShortened {
		count := strings.Count(ipv6, "::")
		if count > 1 {
			return "", fmt.Errorf("multiple double colons found in shortened IPv6 address. Expected only 1")
		}
	}

	ipv6 = ExpandIPv6Address(ipv6)
	segments := strings.Split(ipv6, ":")
	if len(segments) != 8 {
		return "", fmt.Errorf("invalid number of hextets. Found %v expected %v", len(segments), 8)
	}

	hexPattern := regexp.MustCompile("^[0-9a-fA-F]+$")

	for _, segment := range segments {
		isValidHex := hexPattern.MatchString(segment)
		if !isValidHex {
			return "", fmt.Errorf("ipv6 address contains non-hexadecimal hextext %v", segment)
		}

		if len(segment) != 4 {
			return "", fmt.Errorf("expected 4 digit hextet string but found length of %v instead", len(segment))
		}
	}

	return ipv6, nil
}

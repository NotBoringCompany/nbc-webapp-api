package utils

import (
	"strings"

	"golang.org/x/crypto/sha3"
)

/*
ValidChecksum checks whether a given `addr` address is a valid checksum address.
*/
func ValidChecksum(addr string) bool {
	hex := strings.ToLower(addr)[2:]

	d := sha3.NewLegacyKeccak256()
	d.Write([]byte(hex))
	hash := d.Sum(nil)

	ret := "0x"

	for i, b := range hex {
		c := string(b)
		if b < '0' || b > '9' {
			if hash[i/2]&byte(128-i%2*120) != 0 {
				c = string(b - 32)
			}
		}
		ret += c
	}

	return addr == ret
}

package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// TimeFormat represents different time display formats
type TimeFormat int

const (
	TimeFormatNoTime TimeFormat = iota // No time display
	TimeFormatShort                    // Login full (24 chars), logout short (HH:MM)
	TimeFormatFull                     // Full ctime format
	TimeFormatISO                      // ISO8601 format
)

// formatTime formats time according to the specified format
func formatTime(t time.Time, format TimeFormat, isLogin bool) string {
	switch format {
	case TimeFormatNoTime:
		return ""
		
	case TimeFormatShort:
		if isLogin {
			// Full format for login: "Wed Dec 14 14:02" (without seconds)
			return t.Format("Mon Jan 2 15:04")
		}
		// Short format for logout: "14:02"
		return t.Format("15:04")
		
	case TimeFormatFull:
		// Full ctime format: "Wed Dec 14 14:02:09 2025"
		return t.Format(time.ANSIC)
		
	case TimeFormatISO:
		// ISO8601 format: "2025-12-14T14:02:09+0000"
		return t.Format("2006-01-02T15:04:05-0700")
		
	default:
		return t.Format("Mon Jan 2 15:04:05 2006")
	}
}

// performDNSLookup performs reverse DNS lookup on an IP address
func performDNSLookup(addr string) string {
	if addr == "" {
		return ""
	}

	// Check if it's an IPv4-mapped IPv6 address (::ffff:x.x.x.x)
	if strings.HasPrefix(addr, "::ffff:") {
		addr = addr[7:] // Strip the ::ffff: prefix
	}

	// Try to parse as IP
	ip := net.ParseIP(addr)
	if ip == nil {
		// Not an IP address, return as-is
		return addr
	}

	// Perform reverse lookup
	names, err := net.LookupAddr(addr)
	if err != nil || len(names) == 0 {
		return addr
	}

	// Return first hostname, removing trailing dot
	hostname := names[0]
	if strings.HasSuffix(hostname, ".") {
		hostname = hostname[:len(hostname)-1]
	}
	return hostname
}

// parseIPAddress parses an IP address from the AddrV6 field
func parseIPAddress(addrV6 [4]int32) string {
	// Check if it's IPv4 (only first element is non-zero)
	if addrV6[1] == 0 && addrV6[2] == 0 && addrV6[3] == 0 {
		if addrV6[0] == 0 {
			return ""
		}
		// IPv4 address
		b1 := byte(addrV6[0] & 0xFF)
		b2 := byte((addrV6[0] >> 8) & 0xFF)
		b3 := byte((addrV6[0] >> 16) & 0xFF)
		b4 := byte((addrV6[0] >> 24) & 0xFF)
		return fmt.Sprintf("%d.%d.%d.%d", b1, b2, b3, b4)
	}

	// IPv6 address
	// Check if it's IPv4-mapped IPv6
	if addrV6[0] == 0 && addrV6[1] == 0 && uint32(addrV6[2]) == 0xFFFF0000 {
		// IPv4-mapped IPv6: ::ffff:x.x.x.x
		b1 := byte(addrV6[3] & 0xFF)
		b2 := byte((addrV6[3] >> 8) & 0xFF)
		b3 := byte((addrV6[3] >> 16) & 0xFF)
		b4 := byte((addrV6[3] >> 24) & 0xFF)
		return fmt.Sprintf("::ffff:%d.%d.%d.%d", b1, b2, b3, b4)
	}

	// Full IPv6 address
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
		uint16(addrV6[0]&0xFFFF), uint16((addrV6[0]>>16)&0xFFFF),
		uint16(addrV6[1]&0xFFFF), uint16((addrV6[1]>>16)&0xFFFF),
		uint16(addrV6[2]&0xFFFF), uint16((addrV6[2]>>16)&0xFFFF),
		uint16(addrV6[3]&0xFFFF), uint16((addrV6[3]>>16)&0xFFFF))
}

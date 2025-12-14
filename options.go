package main

import (
	"time"
)

// Options holds all command-line options
type Options struct {
	// File options
	WtmpFile string // Path to wtmp/btmp file
	Limit    int    // Number of records to display

	// Display options
	NoHostname bool       // -R/--nohostname: Hide hostname field
	HostLast   bool       // -a/--hostlast: Display hostname in last column
	FullNames  bool       // -w/--fullnames: Display full user and domain names
	ShowSystem bool       // -x/--system: Display system shutdown entries and run level changes
	TimeFormat TimeFormat // Time display format

	// Network options
	DNSLookup bool // -d/--dns: Translate IP to hostname via DNS
	UseIP     bool // -i/--ip: Display IP addresses in numbers-and-dots notation

	// Filter options
	Since      *time.Time // -s/--since: Display lines since specified time
	Until      *time.Time // -t/--until: Display lines until specified time
	Present    *time.Time // -p/--present: Display who were present at specified time
	Users      []string   // Filter by specific users
	Terminals  []string   // Filter by specific terminals
	Hosts      []string   // Filter by specific hosts
}

// NewOptions creates default options
func NewOptions() *Options {
	return &Options{
		WtmpFile:   "/var/log/wtmp",
		TimeFormat: TimeFormatShort,
		Limit:      -1, // No limit by default
	}
}

// ShouldDisplay checks if a record should be displayed based on filters
func (o *Options) ShouldDisplay(record *LoginRecord) bool {
	// Filter by time
	if o.Since != nil && record.LoginTime.Before(*o.Since) {
		return false
	}
	if o.Until != nil && record.LoginTime.After(*o.Until) {
		return false
	}
	if o.Present != nil {
		// Check if user was logged in at the specified time
		if record.LoginTime.After(*o.Present) {
			return false
		}
		if record.LogoutTime != nil && record.LogoutTime.Before(*o.Present) {
			return false
		}
	}

	// Filter by users
	if len(o.Users) > 0 {
		found := false
		for _, user := range o.Users {
			if record.User == user {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by terminals
	if len(o.Terminals) > 0 {
		found := false
		for _, terminal := range o.Terminals {
			if record.Terminal == terminal || record.Terminal == "/dev/"+terminal {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by hosts
	if len(o.Hosts) > 0 {
		found := false
		for _, host := range o.Hosts {
			if record.Host == host {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

package main

import (
	"fmt"
	"strings"
	"time"
)

// LoginRecord represents a processed login session
type LoginRecord struct {
	User      string
	Terminal  string
	Host      string
	LoginTime time.Time
	LogoutTime *time.Time
	Duration   time.Duration
	Status     string // "still logged in", "gone - no logout", "crash", "down", etc.
	PID       int32
}

// LogoutEntry represents a logout record in the linked list
type LogoutEntry struct {
	Terminal   string
	User       string
	LogoutTime time.Time
	Next       *LogoutEntry
}

// LogoutList manages the linked list of logout records
type LogoutList struct {
	Head *LogoutEntry
}

// Add adds a logout entry to the list
func (l *LogoutList) Add(terminal, user string, logoutTime time.Time) {
	entry := &LogoutEntry{
		Terminal:   terminal,
		User:       user,
		LogoutTime: logoutTime,
		Next:       l.Head,
	}
	l.Head = entry
}

// Find finds and removes a matching logout entry
func (l *LogoutList) Find(terminal, user string) *LogoutEntry {
	var prev *LogoutEntry
	current := l.Head

	for current != nil {
		if current.Terminal == terminal && (user == "" || current.User == user) {
			// Found a match, remove from list
			if prev == nil {
				l.Head = current.Next
			} else {
				prev.Next = current.Next
			}
			return current
		}
		prev = current
		current = current.Next
	}

	return nil
}

// Clear clears all logout entries
func (l *LogoutList) Clear() {
	l.Head = nil
}

// ProcessWtmpFile processes wtmp file and returns login records
func ProcessWtmpFile(filename string, options *Options) ([]*LoginRecord, error) {
	records, err := ReadAllRecords(filename)
	if err != nil {
		return nil, err
	}

	var loginRecords []*LoginRecord
	logoutList := &LogoutList{}
	var lastBootTime *time.Time
	var crashTime *time.Time

	// Process records (they are already in reverse order)
	for _, rec := range records {
		switch rec.Type {
		case USER_PROCESS:
			loginRecord := processUserProcess(rec, logoutList, lastBootTime, crashTime, options)
			if loginRecord != nil {
				loginRecords = append(loginRecords, loginRecord)
			}

		case DEAD_PROCESS:
			// Add to logout list for later matching
			terminal := rec.Device()
			user := rec.Username()
			logoutList.Add(terminal, user, rec.Time())

		case BOOT_TIME:
			bootTime := rec.Time()
			lastBootTime = &bootTime
			
			// Clear all logout records before this boot
			logoutList.Clear()
			crashTime = nil

			// Always show boot time (not just with ShowSystem)
			loginRecords = append(loginRecords, &LoginRecord{
				User:      "reboot",
				Terminal:  "system boot",
				Host:      rec.Hostname(),
				LoginTime: bootTime,
				Status:    "still running",
			})

		case SHUTDOWN_TIME:
			shutdownTime := rec.Time()
			crashTime = nil // Not a crash if proper shutdown
			
			if options.ShowSystem {
				loginRecords = append(loginRecords, &LoginRecord{
					User:      "shutdown",
					Terminal:  "system down",
					Host:      "",
					LoginTime: shutdownTime,
				})
			}

		case RUN_LVL:
			if options.ShowSystem {
				runlevel := rec.Hostname()
				loginRecords = append(loginRecords, &LoginRecord{
					User:      "runlevel",
					Terminal:  runlevel,
					Host:      "",
					LoginTime: rec.Time(),
				})
			}

		case OLD_TIME, NEW_TIME:
			if options.ShowSystem {
				what := "old time"
				if rec.Type == NEW_TIME {
					what = "new time"
				}
				loginRecords = append(loginRecords, &LoginRecord{
					User:      what,
					Terminal:  "",
					Host:      "",
					LoginTime: rec.Time(),
				})
			}
		}
	}

	return loginRecords, nil
}

// processUserProcess handles USER_PROCESS records
func processUserProcess(rec *Utmpx, logoutList *LogoutList, lastBootTime, crashTime *time.Time, options *Options) *LoginRecord {
	user := rec.Username()
	terminal := rec.Device()
	host := rec.Hostname()
	loginTime := rec.Time()
	pid := rec.Pid

	// Skip empty records
	if user == "" && terminal == "" {
		return nil
	}

	// Try to find matching logout
	logout := logoutList.Find(terminal, user)
	
	record := &LoginRecord{
		User:      user,
		Terminal:  terminal,
		Host:      host,
		LoginTime: loginTime,
		PID:       pid,
	}

	if logout != nil {
		// Found matching logout
		record.LogoutTime = &logout.LogoutTime
		record.Duration = logout.LogoutTime.Sub(loginTime)
		record.Status = ""
	} else {
		// No logout found - determine status
		if isPhantom(pid, terminal, user) {
			record.Status = "gone - no logout"
		} else if crashTime != nil {
			record.Status = "crash"
			record.LogoutTime = crashTime
			record.Duration = crashTime.Sub(loginTime)
		} else if lastBootTime != nil && loginTime.Before(*lastBootTime) {
			record.Status = "down"
			record.LogoutTime = lastBootTime
			record.Duration = lastBootTime.Sub(loginTime)
		} else {
			record.Status = "still logged in"
		}
	}

	return record
}

// isPhantom checks if a login session is a phantom (orphaned)
func isPhantom(pid int32, terminal, user string) bool {
	if pid <= 0 {
		return false
	}

	// Check /proc/[pid]/loginuid
	// This is a simplified check - in real implementation, we'd verify the UID
	// For now, just check if process exists
	// If process doesn't exist, it's likely a phantom
	
	// Alternatively, check /dev/[tty] ownership
	// In a real implementation, check device ownership
	// For this implementation, we'll return false for simplicity

	return false // Conservative: assume not phantom unless proven
}

// FormatDuration formats duration in "days+hours:mins" format
func FormatDuration(d time.Duration) string {
	if d < 0 {
		return ""
	}

	totalMinutes := int64(d.Minutes())
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	mins := totalMinutes % 60

	if days > 0 {
		return fmt.Sprintf("(%d+%02d:%02d)", days, hours, mins)
	}
	return fmt.Sprintf("(%02d:%02d)", hours, mins)
}

// FormatRecord formats a login record for output
func (r *LoginRecord) Format(options *Options) string {
	var parts []string

	// Username (may be truncated or full based on options)
	username := r.User
	if !options.FullNames && len(username) > 8 {
		username = username[:8]
	}
	parts = append(parts, fmt.Sprintf("%-8s", username))

	// Terminal
	terminal := r.Terminal
	if strings.HasPrefix(terminal, "/dev/") {
		terminal = terminal[5:]
	}
	parts = append(parts, fmt.Sprintf("%-12s", terminal))

	// Hostname (unless suppressed)
	if !options.NoHostname {
		host := r.Host
		if options.UseIP {
			// Keep IP address as-is
		} else if options.DNSLookup {
			// Perform DNS lookup
			host = performDNSLookup(host)
		}
		
		if !options.FullNames && len(host) > 16 {
			host = host[:16]
		}
		
		if options.HostLast {
			// Will add at end
		} else {
			parts = append(parts, fmt.Sprintf("%-16s", host))
		}
	}

	// Login time
	loginTimeStr := formatTime(r.LoginTime, options.TimeFormat, true)
	parts = append(parts, loginTimeStr)

	// Logout time or status
	if r.Status != "" {
		parts = append(parts, r.Status)
	} else if r.LogoutTime != nil {
		logoutTimeStr := formatTime(*r.LogoutTime, options.TimeFormat, false)
		parts = append(parts, logoutTimeStr)
	}

	// Duration
	if r.LogoutTime != nil && r.Status == "" {
		parts = append(parts, FormatDuration(r.Duration))
	}

	// Add hostname at end if requested
	if !options.NoHostname && options.HostLast {
		host := r.Host
		if !options.FullNames && len(host) > 16 {
			host = host[:16]
		}
		parts = append(parts, host)
	}

	return strings.Join(parts, " ")
}

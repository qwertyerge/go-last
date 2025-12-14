package main

import (
	"testing"
	"time"
)

func TestUtmpxSerialization(t *testing.T) {
	// Create a test record
	utmp := &Utmpx{
		Type: USER_PROCESS,
		Pid:  1234,
	}
	copy(utmp.User[:], []byte("testuser"))
	copy(utmp.Line[:], []byte("tty1"))
	copy(utmp.Host[:], []byte("localhost"))
	utmp.Tv.Sec = 1234567890
	utmp.Tv.Usec = 0

	// Serialize
	data := utmp.ToBytes()
	
	// Deserialize
	utmp2 := &Utmpx{}
	err := utmp2.FromBytes(data)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	// Verify
	if utmp2.Type != USER_PROCESS {
		t.Errorf("Type mismatch: got %d, want %d", utmp2.Type, USER_PROCESS)
	}
	if utmp2.Pid != 1234 {
		t.Errorf("PID mismatch: got %d, want %d", utmp2.Pid, 1234)
	}
	if utmp2.Username() != "testuser" {
		t.Errorf("Username mismatch: got %s, want testuser", utmp2.Username())
	}
	if utmp2.Device() != "tty1" {
		t.Errorf("Device mismatch: got %s, want tty1", utmp2.Device())
	}
	if utmp2.Hostname() != "localhost" {
		t.Errorf("Hostname mismatch: got %s, want localhost", utmp2.Hostname())
	}
}

func TestCstringToString(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte("hello\x00world"), "hello"},
		{[]byte("test"), "test"},
		{[]byte(""), ""},
		{[]byte("a\x00\x00\x00"), "a"},
	}

	for _, tt := range tests {
		result := cstringToString(tt.input)
		if result != tt.expected {
			t.Errorf("cstringToString(%v) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Minute, "(00:30)"},
		{2 * time.Hour, "(02:00)"},
		{2*time.Hour + 30*time.Minute, "(02:30)"},
		{25 * time.Hour, "(1+01:00)"},
		{50*time.Hour + 30*time.Minute, "(2+02:30)"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("FormatDuration(%v) = %s, want %s", tt.duration, result, tt.expected)
		}
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2025, 12, 14, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		format   TimeFormat
		isLogin  bool
		expected string
	}{
		{TimeFormatNoTime, true, ""},
		{TimeFormatNoTime, false, ""},
		{TimeFormatShort, true, "Sun Dec 14 14:30"}, // Fixed: Dec 14, 2025 is a Sunday
		{TimeFormatShort, false, "14:30"},
		{TimeFormatISO, true, "2025-12-14T14:30:45+0000"},
	}

	for _, tt := range tests {
		result := formatTime(testTime, tt.format, tt.isLogin)
		if result != tt.expected {
			t.Errorf("formatTime(%v, %d, %v) = %s, want %s", testTime, tt.format, tt.isLogin, result, tt.expected)
		}
	}
}

func TestParseIPAddress(t *testing.T) {
	tests := []struct {
		addrV6   [4]int32
		expected string
	}{
		{[4]int32{0, 0, 0, 0}, ""},
		{[4]int32{0x0100007f, 0, 0, 0}, "127.0.0.1"}, // 127.0.0.1 in network byte order
	}

	for _, tt := range tests {
		result := parseIPAddress(tt.addrV6)
		if result != tt.expected {
			t.Errorf("parseIPAddress(%v) = %s, want %s", tt.addrV6, result, tt.expected)
		}
	}
}

func TestLogoutList(t *testing.T) {
	list := &LogoutList{}
	now := time.Now()

	// Add entries
	list.Add("tty1", "user1", now)
	list.Add("tty2", "user2", now.Add(time.Minute))

	// Find and remove
	entry := list.Find("tty1", "user1")
	if entry == nil {
		t.Error("Failed to find tty1/user1")
	}
	if entry.Terminal != "tty1" {
		t.Errorf("Terminal mismatch: got %s, want tty1", entry.Terminal)
	}

	// Try to find again (should be removed)
	entry2 := list.Find("tty1", "user1")
	if entry2 != nil {
		t.Error("Entry should have been removed")
	}

	// Find second entry
	entry3 := list.Find("tty2", "")
	if entry3 == nil {
		t.Error("Failed to find tty2")
	}

	// Clear all
	list.Clear()
	if list.Head != nil {
		t.Error("List should be empty after Clear()")
	}
}

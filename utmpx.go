package main

import (
	"bytes"
	"encoding/binary"
	"time"
)

// UTMPX record types
const (
	EMPTY         = 0 // Record does not contain valid info
	RUN_LVL       = 1 // Change in system run-level
	BOOT_TIME     = 2 // Time of system boot
	NEW_TIME      = 3 // Time after system clock change
	OLD_TIME      = 4 // Time before system clock change
	INIT_PROCESS  = 5 // Process spawned by init
	LOGIN_PROCESS = 6 // Session leader for user login
	USER_PROCESS  = 7 // Normal process
	DEAD_PROCESS  = 8 // Terminated process
	ACCOUNTING    = 9 // Accounting record
	SHUTDOWN_TIME = 11 // System shutdown time
)

// UTMPX structure sizes for Linux amd64
const (
	UT_LINESIZE = 32
	UT_NAMESIZE = 32
	UT_HOSTSIZE = 256
)

// Utmpx represents a utmpx/wtmp record
type Utmpx struct {
	Type      int32           // Type of record
	Pid       int32           // PID of login process
	Line      [UT_LINESIZE]byte // Device name of tty - "/dev/"
	Id        [4]byte         // Terminal name suffix, or inittab ID
	User      [UT_NAMESIZE]byte // Username
	Host      [UT_HOSTSIZE]byte // Hostname for remote login, or kernel version for run-level messages
	Exit      ExitStatus      // Exit status of a process marked as DEAD_PROCESS
	Session   int32           // Session ID (glibc doesn't use this)
	Tv        Timeval         // Time entry was made
	AddrV6    [4]int32        // Internet address of remote host; IPv4 address uses just addr_v6[0]
	Reserved  [20]byte        // Reserved for future use
}

// ExitStatus holds process exit status
type ExitStatus struct {
	Termination int16 // Process termination status
	Exit        int16 // Process exit status
}

// Timeval represents time value
type Timeval struct {
	Sec  int32 // Seconds
	Usec int32 // Microseconds
}

// Size returns the binary size of Utmpx structure
func (u *Utmpx) Size() int {
	return 384 // Standard size for Linux amd64
}

// ToBytes serializes Utmpx to byte slice
func (u *Utmpx) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, u)
	return buf.Bytes()
}

// FromBytes deserializes byte slice to Utmpx
func (u *Utmpx) FromBytes(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, u)
}

// Time returns the time as time.Time
func (u *Utmpx) Time() time.Time {
	return time.Unix(int64(u.Tv.Sec), int64(u.Tv.Usec)*1000)
}

// Username returns the username as string
func (u *Utmpx) Username() string {
	return cstringToString(u.User[:])
}

// Hostname returns the hostname as string
func (u *Utmpx) Hostname() string {
	return cstringToString(u.Host[:])
}

// Device returns the device/terminal name as string
func (u *Utmpx) Device() string {
	return cstringToString(u.Line[:])
}

// Terminal returns the terminal ID as string
func (u *Utmpx) Terminal() string {
	return cstringToString(u.Id[:])
}

// cstringToString converts C-style null-terminated byte array to Go string
func cstringToString(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		n = len(b)
	}
	return string(b[:n])
}

# go-last

A Go implementation of the util-linux `last` command for viewing login history.

## Features

- **UTMPX Structure Parsing**: Correctly parses binary wtmp/btmp files
- **Efficient File Reading**: Reads wtmp files from end to beginning with 16KB buffering
- **State Machine Processing**: Handles all record types (USER_PROCESS, DEAD_PROCESS, BOOT_TIME, etc.)
- **Login Pairing**: Matches login/logout records and calculates session duration
- **Multiple Time Formats**: Supports short, full, and ISO8601 time formats
- **DNS Lookup**: Resolves IP addresses to hostnames
- **Flexible Filtering**: Filter by user, terminal, hostname, or time range

## Installation

```bash
go install github.com/qwertyerge/go-last@latest
```

Or build from source:

```bash
git clone https://github.com/qwertyerge/go-last
cd go-last
go build -o go-last
```

## Usage

```bash
# Show all login records
go-last

# Use a specific wtmp file
go-last -f /var/log/wtmp

# Show full timestamps
go-last -F

# Show system events (shutdown, runlevel changes)
go-last -x

# Limit output to 10 lines
go-last -n 10

# Filter by user
go-last username

# Filter by terminal
go-last tty1

# Show logins since a specific time
go-last -s "2025-12-14 10:00:00"

# Hide hostname field
go-last -R

# Show hostname in last column
go-last -a

# Display full user and domain names
go-last -w

# Perform DNS lookup for IP addresses
go-last -d

# Display IP addresses in numeric format
go-last -i
```

## Command-Line Options

| Option | Long Option | Description |
|--------|-------------|-------------|
| `-f FILE` | | Use a specific file instead of /var/log/wtmp |
| `-n NUM` | | Limit the number of lines to display |
| `-R` | `--nohostname` | Don't display the hostname field |
| `-a` | `--hostlast` | Display hostname in the last column |
| `-w` | `--fullnames` | Display full user and domain names |
| `-x` | `--system` | Display system shutdown entries and run level changes |
| `-F` | `--fulltimes` | Print full login and logout times and dates |
| `-d` | `--dns` | Translate IP addresses to hostnames via DNS |
| `-i` | `--ip` | Display IP addresses in numeric format |
| `-s TIME` | `--since` | Display logins since the specified time |
| `-t TIME` | `--until` | Display logins until the specified time |
| `-p TIME` | `--present` | Display who were present at the specified time |

## Time Format

Time can be specified in the following formats:
- `YYYY-MM-DD HH:MM:SS`
- `YYYY-MM-DD HH:MM`
- `YYYY-MM-DD`
- `MM/DD/YYYY HH:MM:SS`
- `MM/DD/YYYY`

## Record Types

The program handles the following wtmp record types:

- **USER_PROCESS**: User login
- **DEAD_PROCESS**: User logout
- **BOOT_TIME**: System boot
- **SHUTDOWN_TIME**: System shutdown
- **RUN_LVL**: Run level change
- **OLD_TIME/NEW_TIME**: System time change

## Output Format

```
USER     TERMINAL     HOSTNAME         LOGIN TIME           LOGOUT TIME  DURATION
reboot   system boot  6.11.0-1018-azur Sun Dec 14 14:00     still running
```

Session duration is displayed in the format:
- `(HH:MM)` for sessions less than 24 hours
- `(D+HH:MM)` for sessions spanning multiple days

## Status Messages

- `still logged in` - Session is still active
- `still running` - System is still running (for boot records)
- `gone - no logout` - Process terminated without proper logout
- `crash` - System crashed
- `down` - System was shut down

## Implementation Details

### UTMPX Structure

The program correctly parses the Linux utmpx structure (384 bytes on amd64):
- Handles different byte orders
- Supports IPv4 and IPv6 addresses
- Recognizes IPv4-mapped IPv6 addresses (::ffff:x.x.x.x)

### File Reading

- Reads wtmp file from end to beginning for efficiency
- Uses 16KB buffering for optimal performance
- Correctly handles records spanning buffer boundaries

### Phantom Process Detection

The program detects "phantom" (orphaned) login sessions by:
1. Checking if the process still exists in /proc
2. Verifying /proc/[pid]/loginuid
3. Falling back to /dev/[tty] ownership check

## Testing

Run the test suite:

```bash
go test -v
```

## License

This project is open source and available under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

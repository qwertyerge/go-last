# Examples

## Basic Usage

### View all login records
```bash
./go-last
```

Output:
```
reboot   system boot  6.11.0-1018-azur Sun Dec 14 14:00 still running

wtmp begins Mon Dec 8 22:06:41 2025
```

### Show full timestamps
```bash
./go-last -F
```

Output:
```
reboot   system boot  6.11.0-1018-azur Sun Dec 14 14:00:51 2025 still running
```

### Show system events
```bash
./go-last -x
```

Output:
```
runlevel 6.11.0-1018-azure                  Sun Dec 14 14:01
reboot   system boot  6.11.0-1018-azur Sun Dec 14 14:00 still running
runlevel 6.11.0-1018-azure                  Mon Dec 8 22:06
```

## Filtering

### Filter by user
```bash
./go-last john
```

### Filter by terminal
```bash
./go-last tty1
```

### Show logins since a specific time
```bash
./go-last -s "2025-12-14 10:00:00"
```

### Show logins until a specific time
```bash
./go-last -t "2025-12-14 18:00:00"
```

### Show who was logged in at a specific time
```bash
./go-last -p "2025-12-14 12:00:00"
```

## Display Options

### Hide hostname
```bash
./go-last -R
```

### Show hostname in last column
```bash
./go-last -a
```

### Display full user and domain names
```bash
./go-last -w
```

### Limit output to N lines
```bash
./go-last -n 10
```

## Network Options

### Translate IP addresses to hostnames
```bash
./go-last -d
```

### Display IP addresses in numeric format
```bash
./go-last -i
```

## Using Custom wtmp File

```bash
./go-last -f /var/log/wtmp.1
```

## Combining Options

### Show last 5 system events with full times
```bash
./go-last -x -F -n 5
```

### Show all logins by user 'john' since yesterday
```bash
./go-last -s "2025-12-13 00:00:00" john
```

### Show logins with IP addresses and no hostname field
```bash
./go-last -i -R
```

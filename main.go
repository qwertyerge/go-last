package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	// File options
	fileFlag  = flag.String("f", "", "Use a specific file instead of /var/log/wtmp")
	limitFlag = flag.Int("n", -1, "Limit the number of lines to display")

	// Display options
	noHostnameFlag = flag.Bool("R", false, "Don't display the hostname field")
	hostLastFlag   = flag.Bool("a", false, "Display hostname in the last column")
	fullNamesFlag  = flag.Bool("w", false, "Display full user and domain names")
	showSystemFlag = flag.Bool("x", false, "Display system shutdown entries and run level changes")
	fullTimesFlag  = flag.Bool("F", false, "Print full login and logout times and dates")

	// Network options
	dnsFlag = flag.Bool("d", false, "Translate the IP number back into a hostname")
	ipFlag  = flag.Bool("i", false, "Display IP addresses in numbers-and-dots notation")

	// Time filter options
	sinceFlag   = flag.String("s", "", "Display the state of logins since the specified time (YYYY-MM-DD HH:MM:SS)")
	untilFlag   = flag.String("t", "", "Display the state of logins until the specified time (YYYY-MM-DD HH:MM:SS)")
	presentFlag = flag.String("p", "", "Display who were present at the specified time (YYYY-MM-DD HH:MM:SS)")

	// Alternative flags
	noHostnameLongFlag = flag.Bool("nohostname", false, "Don't display the hostname field")
	hostLastLongFlag   = flag.Bool("hostlast", false, "Display hostname in the last column")
	fullNamesLongFlag  = flag.Bool("fullnames", false, "Display full user and domain names")
	showSystemLongFlag = flag.Bool("system", false, "Display system shutdown entries and run level changes")
	fullTimesLongFlag  = flag.Bool("fulltimes", false, "Print full login and logout times and dates")
	dnsLongFlag        = flag.Bool("dns", false, "Translate the IP number back into a hostname")
	ipLongFlag         = flag.Bool("ip", false, "Display IP addresses in numbers-and-dots notation")
	sinceLongFlag      = flag.String("since", "", "Display the state of logins since the specified time")
	untilLongFlag      = flag.String("until", "", "Display the state of logins until the specified time")
	presentLongFlag    = flag.String("present", "", "Display who were present at the specified time")
)

func main() {
	flag.Parse()

	// Create options
	options := NewOptions()

	// Set file
	if *fileFlag != "" {
		options.WtmpFile = *fileFlag
	}

	// Set limit
	if *limitFlag > 0 {
		options.Limit = *limitFlag
	}

	// Set display options
	options.NoHostname = *noHostnameFlag || *noHostnameLongFlag
	options.HostLast = *hostLastFlag || *hostLastLongFlag
	options.FullNames = *fullNamesFlag || *fullNamesLongFlag
	options.ShowSystem = *showSystemFlag || *showSystemLongFlag

	// Set time format
	if *fullTimesFlag || *fullTimesLongFlag {
		options.TimeFormat = TimeFormatFull
	}

	// Set network options
	options.DNSLookup = *dnsFlag || *dnsLongFlag
	options.UseIP = *ipFlag || *ipLongFlag

	// Parse time filters
	if since := getTimeFlag(*sinceFlag, *sinceLongFlag); since != nil {
		options.Since = since
	}
	if until := getTimeFlag(*untilFlag, *untilLongFlag); until != nil {
		options.Until = until
	}
	if present := getTimeFlag(*presentFlag, *presentLongFlag); present != nil {
		options.Present = present
	}

	// Get user/terminal/host filters from remaining arguments
	args := flag.Args()
	for _, arg := range args {
		if strings.HasPrefix(arg, "tty") || strings.HasPrefix(arg, "pts/") {
			options.Terminals = append(options.Terminals, arg)
		} else {
			options.Users = append(options.Users, arg)
		}
	}

	// Process wtmp file
	records, err := ProcessWtmpFile(options.WtmpFile, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", options.WtmpFile, err)
		os.Exit(1)
	}

	// Display records
	count := 0
	for _, record := range records {
		if !options.ShouldDisplay(record) {
			continue
		}

		fmt.Println(record.Format(options))
		count++

		if options.Limit > 0 && count >= options.Limit {
			break
		}
	}

	// Print footer with file info
	fmt.Printf("\nwtmp begins %s\n", getWtmpBeginTime(options.WtmpFile))
}

// getTimeFlag returns the time value from flags (handles both short and long forms)
func getTimeFlag(short, long string) *time.Time {
	value := short
	if value == "" {
		value = long
	}
	if value == "" {
		return nil
	}

	// Try different time formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"01/02/2006 15:04:05",
		"01/02/2006",
	}

	for _, format := range formats {
		t, err := time.Parse(format, value)
		if err == nil {
			return &t
		}
	}

	fmt.Fprintf(os.Stderr, "Invalid time format: %s\n", value)
	return nil
}

// getWtmpBeginTime returns the time of the first record in wtmp
func getWtmpBeginTime(filename string) string {
	reader, err := NewWtmpReader(filename)
	if err != nil {
		return "unknown"
	}
	defer reader.Close()

	// Read all records to find the oldest (first written)
	var firstTime *time.Time
	for {
		record, err := reader.ReadRecord()
		if err != nil {
			break
		}
		if record != nil {
			t := record.Time()
			firstTime = &t
		}
	}

	if firstTime == nil {
		return "unknown"
	}

	return firstTime.Format("Mon Jan 2 15:04:05 2006")
}

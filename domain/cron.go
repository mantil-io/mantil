package domain

import (
	"regexp"
	"strconv"
	"strings"
)

// Stolen from: https://gist.github.com/ultrasonex/e1fdb8354408a56df91aa4902d17aa6a
var (
	minuteRegex     = regexp.MustCompile(`^([*]|([0]?[0-5]?[0-9]?)|(([*]|([0]?[0-5]?[0-9]?))(\/|\-)([0]?[0-5]?[0-9]?))|(([0]?[0-5]?[0-9]?)((\,)([0]?[0-5]?[0-9]?))*))$`)
	hourRegex       = regexp.MustCompile(`^([*]|[01]?[0-9]|2[0-3]|(([*]|([01]?[0-9]|2[0-3]?))(\/|\-)([01]?[0-9]|2[0-3]?))|(([01]?[0-9]|2[0-3]?)((\,)([01]?[0-9]|2[0-3]?))*))$`)
	dayOfMonthRegex = regexp.MustCompile(`^([*]|[?]|([0-2]?[0-9]|3[0-1])|(([*]|([0-2]?[0-9]|3[0-1]))(\/|\-)([0-2]?[0-9]|3[0-1]))|(([0-2]?[0-9]|3[0-1])((\,)([0-2]?[0-9]|3[0-1]))*))$`)
	monthRegex      = regexp.MustCompile(
		`^([*]|([0]?[0-9]|1[0-2])|(JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)|` +
			`((([*]|([0]?[0-9]|1[0-2])|(JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)))(\/|\-)(([0]?[0-9]|1[0-2])|` +
			`(JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)))|((([0]?[0-9]|1[0-2])|` +
			`(JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC))((\,)(([0]?[0-9]|1[0-2])|` +
			`(JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)))*))$`,
	)
	dayOfWeekRegex = regexp.MustCompile(
		`([*]|[?]|([0]?[1-7])|` +
			`(SUN|MON|TUE|WED|THU|FRI|SAT)|((([0]?[1-7])|` +
			`([*]|(SUN|MON|TUE|WED|THU|FRI|SAT)))(\/|\-|\,|\#)(([0]?[1-7])|` +
			`(SUN|MON|TUE|WED|THU|FRI|SAT)))|((([0]?[1-7])|` +
			`(SUN|MON|TUE|WED|THU|FRI|SAT))((\,)(([0]?[1-7])|` +
			`(SUN|MON|TUE|WED|THU|FRI|SAT)))*))$`,
	)
	yearRegex = regexp.MustCompile(`^([*]|([1-2][01][0-9][0-9])|(([1-2][01][0-9][0-9])(\/|\-)([1-2][01][0-9][0-9]))|(([1-2][01][0-9][0-9])((\,)([1-2][01][0-9][0-9]))*))$`)
)

func ValidateAWSCron(cron string) bool {
	parts := strings.Split(cron, " ")
	if len(parts) != 6 {
		return false
	}
	res := []*regexp.Regexp{
		minuteRegex,
		hourRegex,
		dayOfMonthRegex,
		monthRegex,
		dayOfWeekRegex,
		yearRegex,
	}
	for idx, re := range res {
		if !re.MatchString(parts[idx]) {
			return false
		}
	}
	if !validateDays(parts[2], parts[4]) {
		return false
	}
	if !validateYear(parts[5]) {
		return false
	}
	return true
}

func validateDays(dayOfMonth, dayOfWeek string) bool {
	if dayOfMonth != "?" && dayOfWeek != "?" ||
		dayOfMonth == "?" && dayOfWeek == "?" {
		return false
	}
	if strings.Contains(dayOfWeek, "#") {
		parts := strings.Split(dayOfWeek, "#")
		if len(parts) < 2 {
			return false
		}
		d, err := strconv.Atoi(parts[1])
		if err != nil || d > 5 {
			return false
		}
	}
	return true
}

func validateYear(year string) bool {
	var sep string
	switch {
	case strings.Contains(year, ","):
		sep = ","
	case strings.Contains(year, "-"):
		sep = "-"
	}
	if sep != "" {
		parts := strings.Split(year, sep)
		for _, p := range parts {
			y, err := strconv.Atoi(p)
			if err != nil || (y < 1970 || y > 2199) {
				return false
			}
		}
	}
	return true
}

package service

import (
	"strconv"
	"strings"
	"time"

	"todoapp/internal/model"
)

// QuickAdd is the result of parsing a quick-add line such as "Pay rent tomorrow !high #home @Finance every month".
type QuickAdd struct {
	Title      string
	DueDate    string
	Priority   int
	Tags       []string
	Project    string
	Recurrence string
}

var priorityWords = map[string]int{
	"!none": 0, "!0": 0,
	"!low": 1, "!l": 1, "!1": 1,
	"!medium": 2, "!med": 2, "!m": 2, "!2": 2,
	"!high": 3, "!h": 3, "!3": 3,
}

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "sun": time.Sunday,
	"monday": time.Monday, "mon": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday,
	"wednesday": time.Wednesday, "wed": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday,
	"friday": time.Friday, "fri": time.Friday,
	"saturday": time.Saturday, "sat": time.Saturday,
}

var unitWords = map[string]string{
	"day": model.UnitDaily, "days": model.UnitDaily, "daily": model.UnitDaily,
	"week": model.UnitWeekly, "weeks": model.UnitWeekly, "weekly": model.UnitWeekly,
	"month": model.UnitMonthly, "months": model.UnitMonthly, "monthly": model.UnitMonthly,
	"year": model.UnitYearly, "years": model.UnitYearly, "yearly": model.UnitYearly,
}

// ParseQuickAdd extracts priority, tags, project, due date and recurrence from text.
// today is the current day in the user's time zone. Everything not recognised becomes the title.
func ParseQuickAdd(text string, today time.Time) QuickAdd {
	words := strings.Fields(text)
	var qa QuickAdd
	var title []string

	for i := 0; i < len(words); i++ {
		word := words[i]
		lower := strings.ToLower(word)

		if p, ok := priorityWords[lower]; ok {
			qa.Priority = p
			continue
		}
		if len(word) > 1 && word[0] == '#' {
			qa.Tags = append(qa.Tags, word[1:])
			continue
		}
		if len(word) > 1 && word[0] == '@' {
			qa.Project = strings.ReplaceAll(word[1:], "_", " ")
			continue
		}
		if used, due := parseDate(words[i:], today); used > 0 {
			qa.DueDate = due
			i += used - 1
			continue
		}
		if used, rule := parseRecurrence(words[i:]); used > 0 {
			qa.Recurrence = rule
			i += used - 1
			continue
		}
		title = append(title, word)
	}
	qa.Title = strings.Join(title, " ")
	return qa
}

// parseDate recognises a date phrase at the start of words and returns how many words it used.
func parseDate(words []string, today time.Time) (int, string) {
	w := strings.ToLower(words[0])
	format := func(t time.Time) string { return t.Format(model.DateLayout) }

	switch w {
	case "today":
		return 1, format(today)
	case "tomorrow", "tmr":
		return 1, format(today.AddDate(0, 0, 1))
	}
	if model.ValidDate(w) {
		return 1, w
	}
	if wd, ok := weekdays[w]; ok {
		return 1, format(nextWeekday(today, wd))
	}
	if w == "next" && len(words) > 1 {
		next := strings.ToLower(words[1])
		if next == "week" {
			return 2, format(today.AddDate(0, 0, 7))
		}
		if wd, ok := weekdays[next]; ok {
			return 2, format(nextWeekday(today, wd))
		}
	}
	if w == "in" && len(words) > 2 {
		n, err := strconv.Atoi(words[1])
		unit := strings.ToLower(words[2])
		if err == nil && n > 0 && n <= 3650 {
			switch unit {
			case "day", "days":
				return 3, format(today.AddDate(0, 0, n))
			case "week", "weeks":
				return 3, format(today.AddDate(0, 0, 7*n))
			case "month", "months":
				return 3, format(addMonths(today, n))
			}
		}
	}
	return 0, ""
}

// nextWeekday returns the next occurrence of wd strictly after today.
func nextWeekday(today time.Time, wd time.Weekday) time.Time {
	diff := (int(wd) - int(today.Weekday()) + 7) % 7
	if diff == 0 {
		diff = 7
	}
	return today.AddDate(0, 0, diff)
}

// parseDayList recognises weekday names such as "mon wed fri", "mon,wed,fri", "monday/friday" or "weekday(s)"
// (Mon-Fri) and returns how many words were used.
func parseDayList(words []string) (int, []time.Weekday) {
	set := map[time.Weekday]bool{}
	used := 0
	for _, word := range words {
		parts := strings.FieldsFunc(strings.ToLower(word), func(r rune) bool { return r == ',' || r == '/' })
		if len(parts) == 0 {
			break
		}
		add := map[time.Weekday]bool{}
		for _, p := range parts {
			if p == "weekday" || p == "weekdays" {
				for d := time.Monday; d <= time.Friday; d++ {
					add[d] = true
				}
				continue
			}
			d, ok := weekdays[p]
			if !ok {
				add = nil
				break
			}
			add[d] = true
		}
		if add == nil {
			break
		}
		for d := range add {
			set[d] = true
		}
		used++
	}
	if used == 0 {
		return 0, nil
	}
	return used, model.SortedDays(set)
}

// firstOnOrAfter returns the first day on or after from whose weekday is in days.
func firstOnOrAfter(from time.Time, days []time.Weekday) time.Time {
	for i := 0; i < 7; i++ {
		d := from.AddDate(0, 0, i)
		for _, wd := range days {
			if d.Weekday() == wd {
				return d
			}
		}
	}
	return from
}

// parseRecurrence recognises "daily", "every week", "every 2 months" and returns words used.
func parseRecurrence(words []string) (int, string) {
	w := strings.ToLower(words[0])
	if unit, ok := unitWords[w]; ok && strings.HasSuffix(w, "ly") {
		return 1, model.Recurrence{Unit: unit, Interval: 1}.String()
	}
	if w != "every" || len(words) < 2 {
		return 0, ""
	}
	if used, days := parseDayList(words[1:]); used > 0 {
		return 1 + used, model.Recurrence{Unit: model.UnitWeekly, Interval: 1, Days: days}.String()
	}
	if unit, ok := unitWords[strings.ToLower(words[1])]; ok {
		return 2, model.Recurrence{Unit: unit, Interval: 1}.String()
	}
	if len(words) > 2 {
		n, err := strconv.Atoi(words[1])
		unit, ok := unitWords[strings.ToLower(words[2])]
		if err == nil && ok && n > 0 && n <= 365 {
			return 3, model.Recurrence{Unit: unit, Interval: n}.String()
		}
	}
	return 0, ""
}

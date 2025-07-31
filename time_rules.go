package sval

import (
	"fmt"
	"slices"
	"time"
)

type TimeRuleName = string

const (
	TimeRuleNameMinDate       TimeRuleName = "min_date"
	TimeRuleNameMaxDate       TimeRuleName = "max_date"
	TimeRuleNameFormats       TimeRuleName = "formats"
	TimeRuleNameTimezones     TimeRuleName = "timezones"
	TimeRuleNameBeforeNow     TimeRuleName = "before_now"
	TimeRuleNameAfterNow      TimeRuleName = "after_now"
	TimeRuleNameWorkday       TimeRuleName = "workday"
	TimeRuleNameWeekdays      TimeRuleName = "weekdays"
	TimeRuleNameRelativeRange TimeRuleName = "relative_range"
	TimeRuleNameHolidays      TimeRuleName = "holidays"
	TimeRuleNameMinTime       TimeRuleName = "min_time"
	TimeRuleNameMaxTime       TimeRuleName = "max_time"
	TimeRuleNameBusinessHrs   TimeRuleName = "business_hours"
)

type TimeRules struct {
	BaseRules
	MinDate *time.Time `json:"min_date,omitempty" yaml:"min_date,omitempty"`
	MaxDate *time.Time `json:"max_date,omitempty" yaml:"max_date,omitempty"`
	// Validator by default uses all formats from time package.
	// For custom behavior, you can specify your own formats.
	// For example, if you want to use only RFC3339 format, you can write "rfc3339" and thats it!
	Formats   []string `json:"formats,omitempty" yaml:"formats,omitempty"`
	Timezones []string `json:"timezones,omitempty" yaml:"timezones,omitempty"`

	BeforeNow     bool           `json:"before_now,omitempty" yaml:"before_now,omitempty"`
	AfterNow      bool           `json:"after_now,omitempty" yaml:"after_now,omitempty"`
	RelativeRange *time.Duration `json:"relative_range,omitempty" yaml:"relative_range,omitempty"`

	Workday  bool           `json:"workday,omitempty" yaml:"workday,omitempty"`
	Weekdays []time.Weekday `json:"weekdays,omitempty" yaml:"weekdays,omitempty"`
	Holidays []time.Time    `json:"holidays,omitempty" yaml:"holidays,omitempty"`

	MinTime     *time.Time     `json:"min_time,omitempty" yaml:"min_time,omitempty"`
	MaxTime     *time.Time     `json:"max_time,omitempty" yaml:"max_time,omitempty"`
	BusinessHrs *BusinessHours `json:"business_hours,omitempty" yaml:"business_hours,omitempty"`
}

type BusinessHours struct {
	Start    string         `json:"start" yaml:"start"` // Format: "HH:MM"
	End      string         `json:"end" yaml:"end"`     // Format: "HH:MM"
	Days     []time.Weekday `json:"days" yaml:"days"`
	Timezone string         `json:"timezone" yaml:"timezone"` // e.g. "Europe/Moscow"
}

func parseTimeString(s string, formats []string, timezones []string) (time.Time, error) {
	// TODO: make map with all formats from time package
	defaultFormats := []string{
		time.RFC3339,
	}

	if len(formats) == 0 {
		formats = defaultFormats
	}

	if len(timezones) == 0 {
		timezones = []string{"UTC"}
	}

	var lastErr error
	for _, tz := range timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			lastErr = fmt.Errorf("invalid timezone %q: %w", tz, err)
			continue
		}

		for _, f := range formats {
			if t, err := time.ParseInLocation(f, s, loc); err == nil {
				return t, nil
			} else {
				lastErr = err
			}
		}
	}

	if lastErr != nil {
		return time.Time{}, fmt.Errorf("could not parse time: %w", lastErr)
	}

	return time.Time{}, fmt.Errorf("invalid time format")
}

func parseTimeHM(s string, loc *time.Location) (hours, minutes int, err error) {
	t, err := time.ParseInLocation("15:04", s, loc)
	if err != nil {
		return 0, 0, err
	}
	return t.Hour(), t.Minute(), nil
}

func (r *TimeRules) Validate(i any) error {
	err := NewValidationError()

	if i == nil {
		if r.Required {
			err.AddError(BaseRuleNameRequired, r.Required, nil, FieldIsRequired)
			return err
		}
		return nil
	}

	var t time.Time

	switch v := i.(type) {
	case time.Time:
		t = v
	case *time.Time:
		if v == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, nil, FieldIsRequired)
				return err
			}
			return nil
		}
		t = *v
	case string:

		parsed, parseErr := parseTimeString(v, r.Formats, r.Timezones)
		// TODO: add errors.Is to compare parseErr for better ux
		if parseErr != nil {
			err.AddError(TimeRuleNameFormats, r.Formats, v, "invalid time format or timezone")
			return err
		}
		t = parsed
	case *string:
		if v == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, nil, FieldIsRequired)
				return err
			}
			return nil
		}

		parsed, parseErr := parseTimeString(*v, r.Formats, r.Timezones)
		if parseErr != nil {
			err.AddError(TimeRuleNameFormats, r.Formats, *v, "invalid time format or timezone")
			return err
		}
		t = parsed
	case int64:
		t = time.Unix(v, 0)
	case *int64:
		if v == nil {
			if r.Required {
				err.AddError(BaseRuleNameRequired, r.Required, nil, FieldIsRequired)
				return err
			}
			return nil
		}
		t = time.Unix(*v, 0)

	default:
		err.AddError(BaseRuleNameType, "time.Time/string/int64", i, "value must be a time.Time or string or int64 or ptrs to them")
		return err
	}

	if r.MinDate != nil && t.Before(*r.MinDate) {
		err.AddError(TimeRuleNameMinDate, r.MinDate, t, "date is before minimum allowed date")
	}

	if r.MaxDate != nil && t.After(*r.MaxDate) {
		err.AddError(TimeRuleNameMaxDate, r.MaxDate, t, "date is after maximum allowed date")
	}

	now := time.Now()
	if r.BeforeNow && t.After(now) {
		err.AddError(TimeRuleNameBeforeNow, now, t, "date must be before current time")
	}

	if r.AfterNow && t.Before(now) {
		err.AddError(TimeRuleNameAfterNow, now, t, "date must be after current time")
	}

	if r.RelativeRange != nil {
		min := now.Add(-*r.RelativeRange)
		max := now.Add(*r.RelativeRange)
		if t.Before(min) || t.After(max) {
			err.AddError(TimeRuleNameRelativeRange, r.RelativeRange, t, "date is outside the allowed relative range")
		}
	}

	if r.Workday {
		weekday := t.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			err.AddError(TimeRuleNameWorkday, true, t, "date must be a workday")
		}
	}

	if len(r.Weekdays) > 0 {
		weekday := t.Weekday()

		if !slices.Contains(r.Weekdays, weekday) {
			err.AddError(TimeRuleNameWeekdays, r.Weekdays, t, "date must be on one of the allowed weekdays")
		}
	}

	if len(r.Holidays) > 0 {
		for _, holiday := range r.Holidays {
			if t.Year() == holiday.Year() && t.Month() == holiday.Month() && t.Day() == holiday.Day() {
				err.AddError(TimeRuleNameHolidays, r.Holidays, t, "date cannot be a holiday")
				break
			}
		}
	}

	loc := time.UTC

	if len(r.Timezones) > 0 {
		var tzErr error
		for _, tz := range r.Timezones {
			if loc, tzErr = time.LoadLocation(tz); tzErr == nil {
				break
			}
		}
		if tzErr != nil {
			err.AddError(TimeRuleNameTimezones, r.Timezones, t, "no valid timezone found")
			return err
		}
	}

	tInLoc := t.In(loc)
	timeOnly := time.Date(0, 1, 1, tInLoc.Hour(), tInLoc.Minute(), tInLoc.Second(), tInLoc.Nanosecond(), loc)

	if r.MinTime != nil {
		minTime := time.Date(0, 1, 1, r.MinTime.Hour(), r.MinTime.Minute(), r.MinTime.Second(), r.MinTime.Nanosecond(), loc)
		if timeOnly.Before(minTime) {
			err.AddError(TimeRuleNameMinTime, r.MinTime, t, "time is before minimum allowed time")
		}
	}

	if r.MaxTime != nil {
		maxTime := time.Date(0, 1, 1, r.MaxTime.Hour(), r.MaxTime.Minute(), r.MaxTime.Second(), r.MaxTime.Nanosecond(), loc)
		if timeOnly.After(maxTime) {
			err.AddError(TimeRuleNameMaxTime, r.MaxTime, t, "time is after maximum allowed time")
		}
	}

	if r.BusinessHrs != nil {
		bhLoc := loc
		if r.BusinessHrs.Timezone != "" {
			var tzErr error
			bhLoc, tzErr = time.LoadLocation(r.BusinessHrs.Timezone)
			if tzErr != nil {
				err.AddError(TimeRuleNameTimezones, r.BusinessHrs.Timezone, t, "invalid timezone for business hours")
				return err
			}
		}

		tInBH := t.In(bhLoc)

		if len(r.BusinessHrs.Days) > 0 {
			weekday := tInBH.Weekday()

			if !slices.Contains(r.BusinessHrs.Days, weekday) {
				err.AddError(TimeRuleNameBusinessHrs, r.BusinessHrs, t, "time is not within business days")
				return err
			}
		}

		startHour, startMin, startErr := parseTimeHM(r.BusinessHrs.Start, bhLoc)
		if startErr != nil {
			err.AddError(TimeRuleNameBusinessHrs, r.BusinessHrs.Start, t, "invalid business hours start time format")
			return err
		}

		endHour, endMin, endErr := parseTimeHM(r.BusinessHrs.End, bhLoc)
		if endErr != nil {
			err.AddError(TimeRuleNameBusinessHrs, r.BusinessHrs.End, t, "invalid business hours end time format")
			return err
		}

		timeOnly := time.Date(0, 1, 1, tInBH.Hour(), tInBH.Minute(), 0, 0, bhLoc)
		businessStart := time.Date(0, 1, 1, startHour, startMin, 0, 0, bhLoc)
		businessEnd := time.Date(0, 1, 1, endHour, endMin, 0, 0, bhLoc)

		if timeOnly.Before(businessStart) || timeOnly.After(businessEnd) {
			err.AddError(TimeRuleNameBusinessHrs, r.BusinessHrs, t, "time is not within business hours")
		}
	}

	if err.HasErrors() {
		return err
	}

	return nil
}

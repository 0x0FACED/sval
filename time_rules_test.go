package sval

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeRules(t *testing.T) {
	moscow, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)
	workday := time.Date(2025, 7, 31, 14, 30, 0, 0, time.UTC) // Wednesday (or not, i dont want to check this now, mb later xd)
	weekend := time.Date(2025, 8, 2, 14, 30, 0, 0, time.UTC)  // Saturday (or not, i dont want to check this now, mb later xd)
	businessTime := time.Date(2025, 7, 30, 14, 30, 0, 0, moscow)
	afterHours := time.Date(2025, 7, 31, 22, 30, 0, 0, moscow)

	tests := []struct {
		name    string
		rules   TimeRules
		value   interface{}
		wantErr bool
	}{
		// Basic validation tests
		{
			name:    "nil value when not required",
			rules:   TimeRules{},
			value:   nil,
			wantErr: false,
		},
		{
			name:    "nil value when required",
			rules:   TimeRules{BaseRules: BaseRules{Required: true}},
			value:   nil,
			wantErr: true,
		},
		{
			name:    "invalid type",
			rules:   TimeRules{},
			value:   123.45,
			wantErr: true,
		},

		// Time type tests
		{
			name:    "valid time.Time value",
			rules:   TimeRules{},
			value:   now,
			wantErr: false,
		},
		{
			name:    "valid time.Time pointer",
			rules:   TimeRules{},
			value:   &now,
			wantErr: false,
		},
		{
			name:    "nil time.Time pointer when not required",
			rules:   TimeRules{},
			value:   (*time.Time)(nil),
			wantErr: false,
		},

		// String parsing tests
		{
			name: "valid RFC3339 string",
			rules: TimeRules{
				Formats: []string{time.RFC3339},
			},
			value:   "2025-07-31T14:30:00Z",
			wantErr: false,
		},
		{
			name: "invalid time format",
			rules: TimeRules{
				Formats: []string{time.RFC3339},
			},
			value:   "31.07.2025",
			wantErr: true,
		},
		{
			name: "multiple valid formats",
			rules: TimeRules{
				Formats: []string{time.RFC3339, "2006-01-02"},
			},
			value:   "2025-07-31",
			wantErr: false,
		},

		// Timezone tests
		{
			name: "valid timezone",
			rules: TimeRules{
				Timezones: []string{"Europe/Moscow"},
			},
			value:   now,
			wantErr: false,
		},
		{
			name: "invalid timezone",
			rules: TimeRules{
				Timezones: []string{"Invalid/Zone"},
			},
			value:   now,
			wantErr: true,
		},
		{
			name: "multiple timezones",
			rules: TimeRules{
				Timezones: []string{"Europe/Moscow", "America/New_York"},
			},
			value:   now,
			wantErr: false,
		},

		// MinDate/MaxDate tests
		{
			name: "before min date",
			rules: TimeRules{
				MinDate: &now,
			},
			value:   past,
			wantErr: true,
		},
		{
			name: "after max date",
			rules: TimeRules{
				MaxDate: &now,
			},
			value:   future,
			wantErr: true,
		},
		{
			name: "within date range",
			rules: TimeRules{
				MinDate: &past,
				MaxDate: &future,
			},
			value:   now,
			wantErr: false,
		},

		// BeforeNow/AfterNow tests
		{
			name: "before now check failed",
			rules: TimeRules{
				BeforeNow: true,
			},
			value:   future,
			wantErr: true,
		},
		{
			name: "after now check failed",
			rules: TimeRules{
				AfterNow: true,
			},
			value:   past,
			wantErr: true,
		},

		// RelativeRange tests
		{
			name: "within relative range",
			rules: TimeRules{
				RelativeRange: ptr(12 * time.Hour),
			},
			value:   now.Add(6 * time.Hour),
			wantErr: false,
		},
		{
			name: "outside relative range",
			rules: TimeRules{
				RelativeRange: ptr(12 * time.Hour),
			},
			value:   now.Add(24 * time.Hour),
			wantErr: true,
		},

		// Workday tests
		{
			name: "valid workday",
			rules: TimeRules{
				Workday: true,
			},
			value:   workday,
			wantErr: false,
		},
		{
			name: "weekend not allowed",
			rules: TimeRules{
				Workday: true,
			},
			value:   weekend,
			wantErr: true,
		},

		// Weekdays tests
		{
			name: "allowed weekday",
			rules: TimeRules{
				Weekdays: []time.Weekday{time.Wednesday, time.Thursday},
			},
			value:   workday,
			wantErr: false,
		},
		{
			name: "not allowed weekday",
			rules: TimeRules{
				Weekdays: []time.Weekday{time.Monday, time.Tuesday},
			},
			value:   workday,
			wantErr: true,
		},

		// Business hours tests
		{
			name: "within business hours",
			rules: TimeRules{
				BusinessHrs: &BusinessHours{
					Start:    "09:00",
					End:      "18:00",
					Days:     []time.Weekday{time.Wednesday},
					Timezone: "Europe/Moscow",
				},
			},
			value:   businessTime,
			wantErr: false,
		},
		{
			name: "outside business hours",
			rules: TimeRules{
				BusinessHrs: &BusinessHours{
					Start:    "09:00",
					End:      "18:00",
					Days:     []time.Weekday{time.Wednesday},
					Timezone: "Europe/Moscow",
				},
			},
			value:   afterHours,
			wantErr: true,
		},
		{
			name: "business hours wrong timezone",
			rules: TimeRules{
				BusinessHrs: &BusinessHours{
					Start:    "09:00",
					End:      "18:00",
					Days:     []time.Weekday{time.Wednesday},
					Timezone: "Invalid/Zone",
				},
			},
			value:   businessTime,
			wantErr: true,
		},

		// Combined rules tests
		{
			name: "complex valid case",
			rules: TimeRules{
				Formats:   []string{time.RFC3339},
				Timezones: []string{"Europe/Moscow", "UTC"},
				Workday:   true,
				BusinessHrs: &BusinessHours{
					Start:    "09:00",
					End:      "18:00",
					Days:     []time.Weekday{time.Wednesday, time.Monday, time.Tuesday, time.Thursday, time.Friday},
					Timezone: "Europe/Moscow",
				},
			},
			value:   "2025-07-28T14:30:00+03:00",
			wantErr: false,
		},
		{
			name: "complex invalid case",
			rules: TimeRules{
				Formats:   []string{time.RFC3339},
				Timezones: []string{"Europe/Moscow", "UTC"},
				Workday:   true,
				BusinessHrs: &BusinessHours{
					Start:    "09:00",
					End:      "18:00",
					Days:     []time.Weekday{time.Wednesday},
					Timezone: "Europe/Moscow",
				},
			},
			value:   "2025-08-02T14:30:00+03:00", // Saturday (or not, i dont want to check this now, mb later xd)
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rules.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

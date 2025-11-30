package archmage_test

import (
	"encoding/json/v2"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/shadowopera/sdk-go/archmage"
)

func TestParseDurationShards(t *testing.T) {
	type Trial struct {
		subject  string
		input    any
		expected time.Duration
		expErr   string
	}

	dataset := []Trial{
		{
			subject:  "nil input",
			input:    nil,
			expected: 0,
			expErr:   "",
		},
		{
			subject:  "empty slice",
			input:    []int64{},
			expected: 0,
			expErr:   "",
		},
		{
			subject:  "seconds",
			input:    &[2]int64{0, 5},
			expected: 5 * time.Second,
			expErr:   "",
		},
		{
			subject:  "milliseconds",
			input:    &[2]int64{1, 500},
			expected: 500 * time.Millisecond,
			expErr:   "",
		},
		{
			subject:  "microseconds",
			input:    &[2]int64{2, 750},
			expected: 750 * time.Microsecond,
			expErr:   "",
		},
		{
			subject:  "nanoseconds",
			input:    &[2]int64{3, 250},
			expected: 250 * time.Nanosecond,
			expErr:   "",
		},
		{
			subject:  "seconds with nanoseconds",
			input:    &[3]int64{4, 3, 500000001},
			expected: 3*time.Second + 500000001*time.Nanosecond,
			expErr:   "",
		},
		{
			subject:  "unsupported type 1",
			input:    "invalid",
			expected: 0,
			expErr:   "invalid duration shards type",
		},
		{
			subject:  "unsupported type 2",
			input:    &[4]int64{4, 3, 500000000, 0},
			expected: 0,
			expErr:   "invalid duration shards type",
		},
		{
			subject:  "invalid format 2",
			input:    &[2]int64{5, 100},
			expected: 0,
			expErr:   "invalid duration shards format",
		},
		{
			subject:  "invalid format 3",
			input:    &[3]int64{0, 5, 100},
			expected: 0,
			expErr:   "invalid duration shards format",
		},
		{
			subject:  "unsupported length",
			input:    []int64{4, 3, 500000000, 0},
			expected: 0,
			expErr:   "invalid duration shards length",
		},
		{
			subject:  "slice with seconds format",
			input:    []int64{0, 10},
			expected: 10 * time.Second,
			expErr:   "",
		},
		{
			subject:  "slice with seconds and nanoseconds",
			input:    []int64{4, 2, 750000001},
			expected: 2*time.Second + 750000001*time.Nanosecond,
			expErr:   "",
		},
		{
			subject:  "slice with invalid length",
			input:    []int64{0, 1, 2, 3},
			expected: 0,
			expErr:   "invalid duration shards length",
		},
	}

	var hitCases int
	var bottomSensor int
	defer func() {
		if bottomSensor != 0 {
			if hitCases != 3 {
				t.Fatalf("expected 3 hit cases, got %d", hitCases)
			}
		}
	}()

	for _, tt := range dataset {
		t.Run(tt.subject, func(t *testing.T) {
			r, err := archmage.ParseDurationShards(tt.input)
			if tt.expErr != "" {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.HasPrefix(err.Error(), tt.expErr) {
					t.Fatalf("expected error having prefix %q, got %v", tt.expErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if r != tt.expected {
					t.Fatalf("expected result %q, got %q", tt.expected, r)
				}

				d1 := archmage.Duration{Duration: tt.expected}
				data, err := json.Marshal(d1)
				if err != nil {
					t.Fatalf("failed to marshal Duration: %v", err)
				}
				var d2 archmage.Duration
				if err = json.Unmarshal(data, &d2); err != nil {
					t.Fatalf("failed to unmarshal Duration: %v", err)
				}
				if d2.V() != tt.expected {
					t.Fatalf("expected unmarshaled Duration to be %q, got %q", tt.expected, d2.V())
				}

				data, err = json.Marshal(tt.input)
				if err != nil {
					t.Fatalf("failed to marshal input: %v", err)
				}
				var d3 archmage.Duration
				d3.Duration = time.Millisecond
				if err = json.Unmarshal(data, &d3); err != nil {
					t.Fatalf("failed to unmarshal input into Duration: %v", err)
				}
				if d3.V() != tt.expected {
					t.Fatalf("expected unmarshaled input Duration to be %q, got %q", tt.expected, d3.V())
				}

				switch tt.subject {
				case "empty slice":
					hitCases++
				case "slice with seconds format":
					hitCases++
				case "slice with seconds and nanoseconds":
					hitCases++
				default:
					v := archmage.ShardDuration(r)
					if !reflect.DeepEqual(tt.input, v) {
						t.Fatalf("expected ShardDuration to be %v, got %v", tt.input, v)
					}
				}
			}
		})
	}

	bottomSensor = 1
}

package duration

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    time.Duration
	}{
		{"1 minute", "1m", false, time.Minute},
		{"30 minutes", "30min", false, 30 * time.Minute},
		{"5 minute full", "5minute", false, 5 * time.Minute},
		{"1 hour", "1h", false, time.Hour},
		{"12 hours", "12h", false, 12 * time.Hour},
		{"1 day", "1d", false, 24 * time.Hour},
		{"7 days", "7d", false, 7 * 24 * time.Hour},
		{"1 week", "1w", false, 7 * 24 * time.Hour},
		{"2 weeks", "2w", false, 2 * 7 * 24 * time.Hour},
		{"1 month", "1mo", false, 30 * 24 * time.Hour},
		{"3 month full", "3month", false, 3 * 30 * 24 * time.Hour},
		{"empty string", "", true, 0},
		{"no unit", "123", true, 0},
		{"no number", "h", true, 0},
		{"zero", "0h", true, 0},
		{"negative", "-1h", true, 0},
		{"float", "1.5h", true, 0},
		{"unknown unit", "1x", true, 0},
		{"spaces", "1 h", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q, got nil", tt.input)
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error for %q: %v", tt.input, err)
				return
			}

			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dur  time.Duration
		want string
	}{
		{"minutes", 30 * time.Minute, "30m"},
		{"hours", 2 * time.Hour, "2h"},
		{"days", 3 * 24 * time.Hour, "3d"},
		{"weeks", 2 * 7 * 24 * time.Hour, "2w"},
		{"months", 1 * 30 * 24 * time.Hour, "1mo"},
		{"90 minutes", 90 * time.Minute, "90m"},
		{"36 hours", 36 * time.Hour, "36h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Format(tt.dur)
			if got != tt.want {
				t.Errorf("Format(%v) = %q, want %q", tt.dur, got, tt.want)
			}
		})
	}
}

package mute

import (
	"testing"
	"time"
)

func TestParseCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		text       string
		wantErr    bool
		wantDur    time.Duration
		wantReason string
	}{
		{
			name:    "duration only",
			text:    "/m 1h",
			wantDur: time.Hour,
		},
		{
			name:       "duration with reason",
			text:       "/m 2d spam in chat",
			wantDur:    2 * 24 * time.Hour,
			wantReason: "spam in chat",
		},
		{
			name:    "mute alias",
			text:    "/mute 1w",
			wantDur: 7 * 24 * time.Hour,
		},
		{
			name:    "minutes",
			text:    "/m 30m",
			wantDur: 30 * time.Minute,
		},
		{
			name:    "minutes full",
			text:    "/m 30minute",
			wantDur: 30 * time.Minute,
		},
		{
			name:    "months",
			text:    "/m 1mo",
			wantDur: 30 * 24 * time.Hour,
		},
		{
			name:    "months full",
			text:    "/m 2month",
			wantDur: 2 * 30 * 24 * time.Hour,
		},
		{
			name:    "no arguments",
			text:    "/m",
			wantErr: true,
		},
		{
			name:    "invalid duration format",
			text:    "/m abc",
			wantErr: true,
		},
		{
			name:    "zero duration",
			text:    "/m 0h",
			wantErr: true,
		},
		{
			name:    "negative-like duration",
			text:    "/m -1h",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := parseCommand(tt.text)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if cmd.Duration != tt.wantDur {
				t.Errorf("duration = %v, want %v", cmd.Duration, tt.wantDur)
			}

			if cmd.Reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", cmd.Reason, tt.wantReason)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
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

			got, err := parseDuration(tt.input)

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
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
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

			got := formatDuration(tt.dur)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.dur, got, tt.want)
			}
		})
	}
}

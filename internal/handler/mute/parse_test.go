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

			cmd, err := parseMuteCommand(tt.text)

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

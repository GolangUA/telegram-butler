package mute

import (
	"testing"
	"time"
)

func TestParseCommand(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		text       string
		wantErr    bool
		wantDur    time.Duration
		wantReason string
	}{
		"duration only": {
			text:    "/m 1h",
			wantDur: time.Hour,
		},
		"duration with reason": {
			text:       "/m 2d spam in chat",
			wantDur:    2 * 24 * time.Hour,
			wantReason: "spam in chat",
		},
		"mute alias": {
			text:    "/mute 1w",
			wantDur: 7 * 24 * time.Hour,
		},
		"minutes": {
			text:    "/m 30m",
			wantDur: 30 * time.Minute,
		},
		"minutes full": {
			text:    "/m 30minute",
			wantDur: 30 * time.Minute,
		},
		"months": {
			text:    "/m 1mo",
			wantDur: 30 * 24 * time.Hour,
		},
		"months full": {
			text:    "/m 2month",
			wantDur: 2 * 30 * 24 * time.Hour,
		},
		"no arguments": {
			text:    "/m",
			wantErr: true,
		},
		"invalid duration format": {
			text:    "/m abc",
			wantErr: true,
		},
		"zero duration": {
			text:    "/m 0h",
			wantErr: true,
		},
		"negative-like duration": {
			text:    "/m -1h",
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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

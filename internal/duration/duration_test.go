package duration

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		wantErr bool
		want    time.Duration
	}{
		"1 minute":      {input: "1m", want: time.Minute},
		"30 minutes":    {input: "30min", want: 30 * time.Minute},
		"5 minute full": {input: "5minute", want: 5 * time.Minute},
		"1 hour":        {input: "1h", want: time.Hour},
		"12 hours":      {input: "12h", want: 12 * time.Hour},
		"1 day":         {input: "1d", want: 24 * time.Hour},
		"7 days":        {input: "7d", want: 7 * 24 * time.Hour},
		"1 week":        {input: "1w", want: 7 * 24 * time.Hour},
		"2 weeks":       {input: "2w", want: 2 * 7 * 24 * time.Hour},
		"1 month":       {input: "1mo", want: 30 * 24 * time.Hour},
		"3 month full":  {input: "3month", want: 3 * 30 * 24 * time.Hour},
		"empty string":  {input: "", wantErr: true},
		"no unit":       {input: "123", wantErr: true},
		"no number":     {input: "h", wantErr: true},
		"zero":          {input: "0h", wantErr: true},
		"negative":      {input: "-1h", wantErr: true},
		"float":         {input: "1.5h", wantErr: true},
		"unknown unit":  {input: "1x", wantErr: true},
		"spaces":        {input: "1 h", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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

	tests := map[string]struct {
		dur  time.Duration
		want string
	}{
		"minutes":    {dur: 30 * time.Minute, want: "30m"},
		"hours":      {dur: 2 * time.Hour, want: "2h"},
		"days":       {dur: 3 * 24 * time.Hour, want: "3d"},
		"weeks":      {dur: 2 * 7 * 24 * time.Hour, want: "2w"},
		"months":     {dur: 1 * 30 * 24 * time.Hour, want: "1mo"},
		"90 minutes": {dur: 90 * time.Minute, want: "90m"},
		"36 hours":   {dur: 36 * time.Hour, want: "36h"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Format(tt.dur)
			if got != tt.want {
				t.Errorf("Format(%v) = %q, want %q", tt.dur, got, tt.want)
			}
		})
	}
}

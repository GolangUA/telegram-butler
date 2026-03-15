package logger

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"maps"

	"github.com/fatih/color"
)

type PrettyHandler struct {
	opts  slog.HandlerOptions
	l     *log.Logger
	attrs []slog.Attr
}

func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	return &PrettyHandler{
		opts: *opts,
		l:    log.New(out, "local-dev ", 0),
	}
}

func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return level >= minLevel
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch {
	case r.Level < slog.LevelDebug:
		level = color.HiCyanString("TRACE:")
	case r.Level < slog.LevelInfo:
		level = color.MagentaString(level)
	case r.Level < slog.LevelWarn:
		level = color.BlueString(level)
	case r.Level < slog.LevelError:
		level = color.YellowString(level)
	default:
		level = color.RedString(level)
	}

	fields := make(map[string]any)

	for _, a := range h.attrs {
		resolveAttr(fields, a)
	}

	r.Attrs(func(a slog.Attr) bool {
		resolveAttr(fields, a)
		return true
	})

	b, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.WhiteString(r.Message)

	h.l.Println(timeStr, level, msg, color.CyanString(string(b)))

	return nil
}

// resolveAttr resolves slog.Attr values (including LogValuer and Groups) into a flat map.
func resolveAttr(fields map[string]any, a slog.Attr) {
	a.Value = a.Value.Resolve()

	if a.Value.Kind() == slog.KindGroup {
		group := make(map[string]any)

		for _, ga := range a.Value.Group() {
			resolveAttr(group, ga)
		}

		if a.Key == "" {
			// Inline group (no key) — merge into parent
			maps.Copy(fields, group)
		} else {
			fields[a.Key] = group
		}

		return
	}

	val := a.Value.Any()

	// Handle error interface — serialize as string, not struct
	if err, ok := val.(error); ok {
		fields[a.Key] = err.Error()
		return
	}

	fields[a.Key] = val
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	newAttrs = append(newAttrs, attrs...)

	return &PrettyHandler{
		opts:  h.opts,
		l:     h.l,
		attrs: newAttrs,
	}
}

func (h *PrettyHandler) WithGroup(_ string) slog.Handler {
	return h
}

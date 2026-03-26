package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
)

const (
	ansiMoveUpClearLine = "\033[1A\033[2K\r"
	ansiGray            = "\033[90m"
	ansiReset           = "\033[0m"
)

// DeduplicatingHandler wraps an slog.Handler to merge consecutive identical log lines.
// When a line repeats, it prints a "repeated Nx" indicator using ANSI escape codes
// to overwrite the count in-place. Designed for terminal output only.
type DeduplicatingHandler struct {
	inner slog.Handler
	w     io.Writer

	mu          sync.Mutex
	lastKey     recordKey
	repeatCount int
}

// recordKey identifies a log record for deduplication, ignoring timestamps.
type recordKey struct {
	level   slog.Level
	message string
	attrs   string
}

// NewDeduplicatingHandler creates a handler that merges consecutive identical log lines
// using ANSI terminal escape codes. w is the terminal writer (typically os.Stdout).
func NewDeduplicatingHandler(inner slog.Handler, w io.Writer) *DeduplicatingHandler {
	return &DeduplicatingHandler{
		inner: inner,
		w:     w,
	}
}

func (h *DeduplicatingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *DeduplicatingHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := makeRecordKey(r)

	if key == h.lastKey {
		h.repeatCount++
		return h.writeRepeatLine()
	}

	// New/different record — reset and pass through to inner handler.
	h.lastKey = key
	h.repeatCount = 0
	return h.inner.Handle(ctx, r)
}

// WithAttrs delegates to the inner handler.
func (h *DeduplicatingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.inner.WithAttrs(attrs)
}

// WithGroup delegates to the inner handler.
func (h *DeduplicatingHandler) WithGroup(name string) slog.Handler {
	return h.inner.WithGroup(name)
}

// writeRepeatLine writes or overwrites the "repeated Nx" indicator line.
func (h *DeduplicatingHandler) writeRepeatLine() error {
	var line string
	if h.repeatCount > 1 {
		// Overwrite the previous "repeated" line.
		line = fmt.Sprintf("%s%s  ↳ repeated %dx%s\n", ansiMoveUpClearLine, ansiGray, h.repeatCount, ansiReset)
	} else {
		// First repeat — write a new line.
		line = fmt.Sprintf("%s  ↳ repeated %dx%s\n", ansiGray, h.repeatCount, ansiReset)
	}
	_, err := h.w.Write([]byte(line))
	return err
}

// makeRecordKey builds a dedup key from a record, ignoring the timestamp.
func makeRecordKey(r slog.Record) recordKey {
	return recordKey{
		level:   r.Level,
		message: r.Message,
		attrs:   canonicalAttrs(r),
	}
}

type attrPair struct {
	key string
	val string
}

// canonicalAttrs returns a canonical string representation of a record's attributes.
func canonicalAttrs(r slog.Record) string {
	if r.NumAttrs() == 0 {
		return ""
	}

	pairs := make([]attrPair, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		collectAttr("", a, &pairs)
		return true
	})

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	var b strings.Builder
	for i, p := range pairs {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(p.key)
		b.WriteByte('=')
		b.WriteString(p.val)
	}
	return b.String()
}

// collectAttr appends flattened key-value pairs directly into dst,
// using dot-separated keys for nested groups.
func collectAttr(prefix string, a slog.Attr, dst *[]attrPair) {
	key := a.Key
	if prefix != "" {
		key = prefix + "." + key
	}

	if a.Value.Kind() == slog.KindGroup {
		for _, ga := range a.Value.Group() {
			collectAttr(key, ga, dst)
		}
		return
	}

	*dst = append(*dst, attrPair{key: key, val: a.Value.String()})
}

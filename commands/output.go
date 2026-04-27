package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/cli"
)

// Marker is the bracketed status marker that prefixes a task line. The
// concrete strings (without brackets) match the conventions documented
// in issue #202: apply uses ok/changed/skipped/error, plan uses
// ok/+/~/-/!.
type Marker string

const (
	// Apply markers
	MarkerOK      Marker = "ok"
	MarkerChanged Marker = "changed"
	MarkerSkipped Marker = "skipped"
	MarkerError   Marker = "error"

	// Plan markers (PlanStatusOK reuses MarkerOK)
	MarkerCreate     Marker = "+"
	MarkerModify     Marker = "~"
	MarkerDestroy    Marker = "-"
	MarkerProbeError Marker = "!"
)

// markerWidth is the fixed column width that every bracketed marker is
// padded to. Wide enough for the longest marker `[changed]` (9 chars)
// plus a trailing space.
const markerWidth = 10

// continuationIndent matches markerWidth so continuation lines align
// under the task name column.
var continuationIndent = strings.Repeat(" ", markerWidth)

// ApplyCounts holds the running totals printed by the apply summary
// line.
type ApplyCounts struct {
	Tasks   int
	Changed int
	OK      int
	Skipped int
	Errors  int
}

// PlanCounts holds the running totals printed by the plan summary line.
type PlanCounts struct {
	Tasks       int
	WouldChange int
	InSync      int
	Errors      int
}

// Formatter renders the structured per-task output for `apply` and
// `plan`. It owns marker padding, color decisions, and the play /
// summary line shapes so both subcommands stay visually consistent.
//
// Color is resolved once at construction and honours NO_COLOR plus
// stdout-isatty, matching fatih/color's defaults. When the active Ui
// is not the real terminal Ui (e.g. cli.MockUi in tests), color is
// also forced off so test assertions stay stable.
type Formatter struct {
	ui      cli.Ui
	verbose bool
	color   bool
	paint   map[Marker]*color.Color
}

// NewFormatter constructs a Formatter bound to the given Ui. verbose
// controls whether `→`-prefixed continuation lines are emitted under
// each task line in apply mode.
func NewFormatter(ui cli.Ui, verbose bool) *Formatter {
	useColor := !noColorDefault()
	return &Formatter{
		ui:      ui,
		verbose: verbose,
		color:   useColor,
		paint: map[Marker]*color.Color{
			MarkerOK:         color.New(color.FgGreen),
			MarkerChanged:    color.New(color.FgYellow),
			MarkerSkipped:    color.New(color.Faint),
			MarkerError:      color.New(color.FgRed),
			MarkerCreate:     color.New(color.FgGreen),
			MarkerModify:     color.New(color.FgYellow),
			MarkerDestroy:    color.New(color.FgRed),
			MarkerProbeError: color.New(color.FgRed),
		},
	}
}

// Verbose reports whether the formatter is in verbose mode.
func (f *Formatter) Verbose() bool { return f.verbose }

// PlayHeader emits a `==> Play: <name>` line above the per-task
// listing. Used once per play; until #208 lands, callers emit one
// header per run.
func (f *Formatter) PlayHeader(name string) {
	f.ui.Output(fmt.Sprintf("==> Play: %s", name))
}

// TaskLine emits one structured task line: a colored, bracketed,
// padded marker followed by the task name. suffix, when non-empty, is
// appended after two spaces (matching the legacy plan-line layout).
//
// Errored task lines are routed through Ui.Error so they land on stderr
// and inherit the cli-skeleton error styling.
func (f *Formatter) TaskLine(m Marker, name, suffix string) {
	marker := f.paintMarker(m)
	line := marker + name
	if suffix != "" {
		line = line + "  " + suffix
	}
	if m == MarkerError || m == MarkerProbeError {
		f.ui.Error(line)
		return
	}
	f.ui.Output(line)
}

// Continuation emits an indented continuation line under the most
// recent task line. prefix is the leading rune (`!` for errors,
// `→` for the verbose command echo, `-` for plan mutation items).
// Each line of `body` is emitted as its own continuation so multi-line
// stderr renders cleanly under the marker column.
func (f *Formatter) Continuation(prefix rune, body string) {
	if body == "" {
		return
	}
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		out := fmt.Sprintf("%s%c %s", continuationIndent, prefix, line)
		f.ui.Output(out)
	}
}

// ApplySummary emits the blank line and `Summary: ...` footer that
// follows an apply run. elapsed is rendered with one decimal of
// seconds, matching the spec in issue #202.
func (f *Formatter) ApplySummary(c ApplyCounts, elapsed time.Duration) {
	errWord := "errors"
	if c.Errors == 1 {
		errWord = "error"
	}
	f.ui.Output("")
	f.ui.Output(fmt.Sprintf(
		"Summary: %d tasks · %d changed · %d ok · %d skipped · %d %s  (took %.1fs)",
		c.Tasks, c.Changed, c.OK, c.Skipped, c.Errors, errWord, elapsed.Seconds(),
	))
}

// PlanSummary emits the blank line and `Plan: ...` footer that
// follows a plan run. The format intentionally matches the legacy
// shape so existing CI consumers (and the bats partial-match
// assertions in tests/bats/plan.bats) keep working.
func (f *Formatter) PlanSummary(c PlanCounts) {
	f.ui.Output("")
	f.ui.Output(fmt.Sprintf(
		"Plan: %d task(s); %d would change, %d in sync, %d error(s).",
		c.Tasks, c.WouldChange, c.InSync, c.Errors,
	))
}

// paintMarker pads the bracketed marker text to markerWidth and applies
// color when enabled.
func (f *Formatter) paintMarker(m Marker) string {
	text := fmt.Sprintf("[%s]", string(m))
	pad := markerWidth - len(text)
	if pad < 1 {
		pad = 1
	}
	padded := text + strings.Repeat(" ", pad)
	if !f.color {
		return padded
	}
	if c, ok := f.paint[m]; ok {
		// Only color the bracketed text; keep trailing whitespace
		// uncolored so the indent is unaffected by ANSI resets.
		return c.Sprint(text) + strings.Repeat(" ", pad)
	}
	return padded
}


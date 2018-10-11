package logging

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

// NewTerminalLogger creates a new TerminalLogger instance using the specified
// MemoryLogger instances.
func NewTerminalLogger(logs []*MemoryLogger) *TerminalLogger {
	return &TerminalLogger{
		logs:    logs,
		enabled: []bool{true, true, true, true},
		appName: "Horde",
		mutex:   sync.Mutex{},
	}
}

// TerminalLogger is a logger that creates a console logging screen with logs
// that can be toggled runtime.
type TerminalLogger struct {
	logs    []*MemoryLogger
	enabled []bool
	appName string
	mutex   sync.Mutex
}

// Start launches the logging screen. If there's an error launching the screen
// it will be returned. The method doesn't return until the user presses the
// escape key.
func (t *TerminalLogger) Start() error {
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	t.draw()

	quit := make(chan bool)
	go func() {
		counters := make([]int, 4)
		for {
			redraw := false
			for i := 0; i < 4; i++ {
				if t.logs[i].NumEntries() > counters[i] {
					counters[i] = t.logs[i].NumEntries()
					redraw = true
				}
			}
			select {
			case <-quit:
				return
			case <-time.After(75 * time.Millisecond):
				// ok
			}
			if redraw {
				t.draw()
			}
		}
	}()
	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyEsc:
				quit <- true
				return nil
			case termbox.KeyCtrlD:
				t.toggle(DebugLevel)
			case termbox.KeyCtrlI:
				t.toggle(InfoLevel)
			case termbox.KeyCtrlW:
				t.toggle(WarningLevel)
			case termbox.KeyCtrlE:
				t.toggle(ErrorLevel)
			}
		}
		t.draw()
	}
}

// toggle log levels on and off
func (t *TerminalLogger) toggle(level uint) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.enabled[level] = !t.enabled[level]
}

// Draw a string to the screen
func (t *TerminalLogger) drawString(x, y, w int, text string, fg, bg termbox.Attribute) {
	pos := x
	for _, ch := range text {
		termbox.SetCell(pos, y, ch, fg, bg)
		pos++
		if pos > w {
			return
		}
	}
}

// Draw the title bar
func (t *TerminalLogger) drawTitleBar(w int) {
	caption := fmt.Sprintf("%s logs", t.appName)
	xpos := w/2 + len(caption)/2
	title := fmt.Sprintf("%s%s%s", strings.Repeat(" ", xpos), caption, strings.Repeat(" ", w-xpos-len(t.appName)))
	t.drawString(0, 0, w, title, termbox.ColorYellow|termbox.AttrBold, termbox.ColorBlue)
}

// Draw log indicator at the bottom right hand corner
func (t *TerminalLogger) drawIndicator(w, h, pos int, name string, enabled bool, fg, bg termbox.Attribute) {
	x := w - (pos * 5)
	y := h - 1
	if enabled {
		t.drawString(x, y, w, fmt.Sprintf("  %s  ", name), fg, bg)
		return
	}
	t.drawString(x, y, w, fmt.Sprintf("  %s  ", name), termbox.ColorBlue, termbox.ColorBlack)

}

// Draw the status bar
func (t *TerminalLogger) drawStatusBar(w, h int) {
	helpStr := fmt.Sprintf("Ctrl+D, I, W, E: Toggle levels (E:%d/W:%d/I:%d/D:%d)",
		t.logs[ErrorLevel].NumEntries(),
		t.logs[WarningLevel].NumEntries(),
		t.logs[InfoLevel].NumEntries(),
		t.logs[DebugLevel].NumEntries())
	t.drawString(0, h-1, w, strings.Repeat(" ", w), termbox.ColorYellow, termbox.ColorBlue)
	t.drawString(1, h-1, w, helpStr, termbox.ColorYellow, termbox.ColorBlue)

	t.drawIndicator(w, h, 1, "E", t.enabled[ErrorLevel], termbox.ColorWhite, termbox.ColorRed)
	t.drawIndicator(w, h, 2, "W", t.enabled[WarningLevel], termbox.ColorBlack, termbox.ColorYellow)
	t.drawIndicator(w, h, 3, "I", t.enabled[InfoLevel], termbox.ColorBlack, termbox.ColorCyan)
	t.drawIndicator(w, h, 4, "D", t.enabled[DebugLevel], termbox.ColorBlack, termbox.ColorWhite)
}

// Draw the log entries
func (t *TerminalLogger) drawLogs(w, h int) {
	enabled := []*MemoryLogger{}
	for i := 0; i < 4; i++ {
		if t.enabled[i] {
			enabled = append(enabled, t.logs[i])
		}
	}
	if len(enabled) == 0 {
		return
	}
	elems := enabled[0].Merge(enabled[1:]...)
	index := len(elems) - 1
	for i := h - 2; i > 0; i-- {
		if index > -1 {
			msg := fmt.Sprintf("%8s  %-15s  %s", elems[index].Time.Format("15:04:05"), elems[index].Location, elems[index].Message)
			fg := termbox.ColorWhite
			bg := termbox.ColorBlack
			switch elems[index].Level {
			case DebugLevel:
				bg = termbox.ColorBlack
				fg = termbox.ColorWhite
			case InfoLevel:
				bg = termbox.ColorBlack
				fg = termbox.ColorBlue | termbox.AttrBold
			case WarningLevel:
				bg = termbox.ColorBlack
				fg = termbox.ColorYellow | termbox.AttrBold
			case ErrorLevel:
				bg = termbox.ColorBlack
				fg = termbox.ColorRed | termbox.AttrBold
			}
			pad := w - len(msg)
			if pad < 0 {
				pad = 0
			}
			t.drawString(0, i, w, msg+strings.Repeat(" ", pad), fg, bg)
			index--
		}
	}
}

// Redraw the screen
func (t *TerminalLogger) draw() {
	t.mutex.Lock()
	w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	t.drawTitleBar(w)
	t.drawStatusBar(w, h)

	t.drawLogs(w, h)
	termbox.Flush()
	t.mutex.Unlock()
}

package logging

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nsf/termbox-go"
)

func TestTermLogger(t *testing.T) {
	logs := NewMemoryLoggers(100)
	EnableMemoryLogger(logs)
	SetLogLevel(DebugLevel)
	term := NewTerminalLogger(logs)

	Debug(strings.Repeat("aaaaa", 20) + strings.Repeat("bbbbb", 20) + strings.Repeat("ccccc", 20) + strings.Repeat("dddd", 20))
	Info(strings.Repeat("X", 512))
	Warning(strings.Repeat("X", 512))
	Error(strings.Repeat("X", 512))

	go func() {
		time.Sleep(200 * time.Millisecond)
		termbox.Interrupt()
	}()

	// This might fail if there's no TTY configured but for tests
	// on the command line the test will be executed
	term.Start()
}

func TestSplitAndPad(t *testing.T) {
	testSplits := func(str string, max int, expected int) {
		split := splitAndPadLines(str, max)
		if len(split) != expected {
			t.Fatalf("Expected %d lines but got %d: %v (string: %s)", expected, len(split), split, str)
		}
		for i := range split {
			if len(split[i]) != max {
				t.Fatalf("Line %d isn't padded correctly. Expected %d but got %d: %v", i, max, len(split[i]), split)
			}
			for _, ch := range split[i] {
				if ch != ('0'+rune(i)) && ch != ' ' {
					t.Fatalf("Line %d is incorrect. Expected chars '%c' or ' ' but got \"%s\"", i, ('0' + rune(i)), split[i])
				}
			}
		}
		// Each line contains either the line number or a space
	}
	testStr := ""
	for i := 0; i < 10; i++ {
		testStr = testStr + strings.Repeat(fmt.Sprintf("%d", i), 10)
	}
	testSplits("", 10, 1)
	testSplits("0", 10, 1)
	testSplits(testStr[:11], 10, 2)
	testSplits(testStr[:25], 10, 3)
	testSplits(testStr[:30], 10, 3)
	testSplits(testStr[:99], 10, 10)
}

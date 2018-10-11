package logging

import (
	"testing"
	"time"
)

func TestMemlogger(t *testing.T) {
	ml := NewMemoryLogger(10, DebugLevel)
	if len(ml.Entries()) != 0 {
		t.Fatal("Expected 0 elements")
	}
	for i := 0; i < 100; i++ {
		ml.Write([]byte("main.go:57: This is a log entry:with colon:"))
	}
	if len(ml.Entries()) != ml.maxEntries {
		t.Fatal("Expected ", ml.maxEntries, " but got ", len(ml.Entries()))
	}

	ml = NewMemoryLogger(0, ErrorLevel)
	for i := 0; i < 100; i++ {
		ml.Write([]byte("main.go:57: This is a log entry:with colon:"))
	}
}

func TestMemloggerMerge(t *testing.T) {
	m1 := NewMemoryLogger(10, DebugLevel)
	m2 := NewMemoryLogger(8, DebugLevel)
	m3 := NewMemoryLogger(12, DebugLevel)
	m4 := NewMemoryLogger(10, DebugLevel)
	m1.Merge(m2, m3, m4)
	for i := 0; i < 12; i++ {
		m1.Write([]byte("Log entry"))
		m2.Write([]byte("Log entry"))
		m3.Write([]byte("Log entry"))
	}
	entries := m1.Merge(m2, m3, m4)
	prev := time.Now().Add(-time.Second)
	for _, v := range entries {
		if v.Time.Before(prev) {
			t.Fatal("Element should not be before previous")
		}
		prev = v.Time
	}
}
func BenchmarkMemlogger(b *testing.B) {
	ml := NewMemoryLogger(1000, WarningLevel)
	for i := 0; i < b.N; i++ {
		ml.Write([]byte("main.go:57: This is a log entry:with:colon"))
	}
}

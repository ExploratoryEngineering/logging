package logging

import (
	"log"
	"strings"
	"sync"
	"time"
)

// MemoryLoggerFlags is the assumed layout for the log
const MemoryLoggerFlags = log.Lshortfile

// LogEntry is a linked list entry that holds log entries
type LogEntry struct {
	Time    time.Time
	File    string
	Line    string
	Message string
	Next    *LogEntry
}

// NewLogEntry creates a new log entry
func NewLogEntry(input string) *LogEntry {
	fields := strings.Split(input, ":")
	if len(fields) > 2 {
		file := fields[0]
		line := fields[1]
		message := strings.Join(fields[2:], ":")
		return &LogEntry{Time: time.Now(), Message: message, File: file, Line: line, Next: nil}
	}
	return &LogEntry{Time: time.Now(), Message: input, File: "-", Line: "-", Next: nil}
}

// MemoryLogger is a type that logs to memory. The logs are stored in a linked
// list.
type MemoryLogger struct {
	FirstEntry *LogEntry
	LastEntry  *LogEntry
	NumEntries int
	MaxEntries int
	mutex      sync.Mutex
}

// NewMemoryLogger creates a new memory logger
func NewMemoryLogger(maxEntries int) *MemoryLogger {
	if maxEntries < 1 {
		maxEntries = 1
	}
	return &MemoryLogger{
		mutex:      sync.Mutex{},
		MaxEntries: maxEntries,
		FirstEntry: nil,
		LastEntry:  nil,
		NumEntries: 0}
}

func (m *MemoryLogger) addEntry(entry *LogEntry) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.NumEntries++
	e := m.FirstEntry
	if e == nil {
		m.FirstEntry = entry
		m.LastEntry = entry
		return
	}
	m.LastEntry.Next = entry
	m.LastEntry = m.LastEntry.Next

	// Remove the first entry if we exceed max number of entries
	if m.NumEntries > m.MaxEntries {
		remove := m.FirstEntry
		m.FirstEntry = m.FirstEntry.Next
		remove.Next = nil
		return
	}
}

// Write is a stub. This is the io.Writer implementation
func (m *MemoryLogger) Write(p []byte) (n int, err error) {
	m.addEntry(NewLogEntry(string(p)))
	return len(p), nil
}

// Entries returns the entries
func (m *MemoryLogger) Entries() []LogEntry {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var ret []LogEntry
	e := m.FirstEntry
	for e != nil {
		ret = append(ret, *e)
		e = e.Next
	}
	return ret
}

// Merge merges this and a number of other logs
func (m *MemoryLogger) Merge(other ...*MemoryLogger) []LogEntry {
	ret := make([]LogEntry, 0)
	elems := make([]*LogEntry, len(other)+1)
	elems[0] = m.FirstEntry
	m.mutex.Lock()
	for i, o := range other {
		o.mutex.Lock()
		elems[i+1] = o.FirstEntry
	}

	for len(elems) > 0 {
		// Remove all elems that are nil
		remove := make([]int, 0)
		for i := range elems {
			if elems[i] == nil {
				//		elems = append(elems[0:i], elems[i+1:]...)
				remove = append(remove, i)
			}
		}
		for _, i := range remove {
			elems = append(elems[0:i], elems[i+1:]...)
		}
		// Find the next lowest element and move forward
		min := time.Now()
		smallest := -1
		for i := range elems {
			if elems[i].Time.Before(min) {
				min = elems[i].Time
				smallest = i
			}
		}
		if smallest >= 0 {
			ret = append(ret, *elems[smallest])
			elems[smallest] = elems[smallest].Next
		}
	}
	m.mutex.Unlock()
	for _, o := range other {
		o.mutex.Unlock()
	}
	return ret
}

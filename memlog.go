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
	Time     time.Time
	Location string
	Message  string
	Next     *LogEntry
	Level    uint
}

// NewLogEntry creates a new log entry
func NewLogEntry(input string, level uint) *LogEntry {
	fields := strings.Split(input, ":")
	if len(fields) > 2 {
		file := fields[0]
		line := fields[1]
		message := strings.Join(fields[2:], ":")
		return &LogEntry{Time: time.Now(), Message: message, Location: file + ":" + line, Next: nil, Level: level}
	}
	return &LogEntry{Time: time.Now(), Message: input, Location: "-", Next: nil, Level: level}
}

// MemoryLogger is a type that logs to memory. The logs are stored in a linked
// list.
type MemoryLogger struct {
	FirstEntry *LogEntry
	LastEntry  *LogEntry
	numEntries int
	maxEntries int
	level      uint
	mutex      sync.Mutex
}

// NewMemoryLogger creates a new memory logger
func NewMemoryLogger(maxEntries int, l uint) *MemoryLogger {
	if maxEntries < 1 {
		maxEntries = 1
	}
	return &MemoryLogger{
		mutex:      sync.Mutex{},
		maxEntries: maxEntries,
		FirstEntry: nil,
		LastEntry:  nil,
		level:      l,
		numEntries: 0}
}

// NewMemoryLoggers is a convenience function to create logs for all levels
func NewMemoryLoggers(maxEntries int) []*MemoryLogger {
	return []*MemoryLogger{
		NewMemoryLogger(maxEntries, DebugLevel),
		NewMemoryLogger(maxEntries, InfoLevel),
		NewMemoryLogger(maxEntries, WarningLevel),
		NewMemoryLogger(maxEntries, ErrorLevel),
	}
}

func (m *MemoryLogger) addEntry(entry *LogEntry) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.numEntries++
	e := m.FirstEntry
	if e == nil {
		m.FirstEntry = entry
		m.LastEntry = entry
		return
	}
	m.LastEntry.Next = entry
	m.LastEntry = m.LastEntry.Next

	// Remove the first entry if we exceed max number of entries
	if m.numEntries > m.maxEntries {
		remove := m.FirstEntry
		m.FirstEntry = m.FirstEntry.Next
		remove.Next = nil
		return
	}
}

// Write is a stub. This is the io.Writer implementation
func (m *MemoryLogger) Write(p []byte) (n int, err error) {
	m.addEntry(NewLogEntry(string(p), m.level))
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
		newElems := make([]*LogEntry, 0)
		for _, v := range elems {
			if v != nil {
				newElems = append(newElems, v)
			}
		}
		elems = newElems
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

// NumEntries returns the number of entries in the log in total. This
// is increased for every time something is logged to the logger.
func (m *MemoryLogger) NumEntries() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.numEntries
}

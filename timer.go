package main

import (
	"fmt"
	"time"
)

// TimeEntry  represents how long it took to do a specific task
type TimeEntry struct {
	title      string
	start      time.Time
	duration   time.Duration
	subEntries []*TimeEntry
	parent     *TimeEntry
	depth      int
}

// Print prints out information about the time entry and it's sub entries
func (te TimeEntry) Print() {
	for i := 0; i < te.depth; i++ {
		print("---")
	}
	println(fmt.Sprintf("-> %s took %s", te.title, te.duration))
	for _, childEntry := range te.subEntries {
		childEntry.Print()
	}
}

// Timer is meant for timing different parts of the application
type Timer struct {
	lastEntry *TimeEntry
	curDepth  int
}

func (t *Timer) begin(title string) {
	t.lastEntry = &TimeEntry{
		title:  title,
		depth:  t.curDepth,
		start:  time.Now(),
		parent: t.lastEntry,
	}
	t.curDepth++
}

func (t *Timer) end() {
	t.curDepth--
	t.lastEntry.duration = time.Since(t.lastEntry.start)

	if t.lastEntry.parent != nil {
		t.lastEntry.parent.subEntries = append(t.lastEntry.parent.subEntries, t.lastEntry)
	}

	if t.curDepth == 0 {
		t.lastEntry.Print()
	}

	t.lastEntry = t.lastEntry.parent
}

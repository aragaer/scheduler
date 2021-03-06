// -*- tab-width:4  -*-
package scheduler

import (
	"strings"
	"testing"
)

type event struct {
	delay, repeat int64
	name          string
}

func (e event) mk() *Event {
	return &Event{e.delay, e.repeat, e.name, []byte(`{"event": "tick"}`)}
}

type Events []event
type Expected []string

type tc struct {
	name     string
	events   Events
	expected Expected
}

var testCases = []tc{
	{"nothing",
		Events{},
		Expected{"", "", ""}},
	{"one immediate",
		Events{{0, 0, "tick"}},
		Expected{"tick", "", ""}},
	{"one delayed",
		Events{{1, 0, "tick"}},
		Expected{"", "tick", ""}},
	{"one repeating",
		Events{{0, 1, "tick"}},
		Expected{"tick", "tick", "tick"}},
	{"two immediate",
		Events{{0, 0, "tick"}, {0, 0, "tock"}},
		Expected{"tick tock", "", ""}},
	{"two alternating",
		Events{{0, 2, "tick"}, {1, 2, "tock"}},
		Expected{"tick", "tock", "tick", "tock"}},
	{"two different freq",
		Events{{0, 2, "tick"}, {0, 3, "tock"}},
		Expected{"tick tock", "", "tick", "tock", "tick", "", "tick tock"}},
	{"insert before",
		Events{{1, 0, "tock"}, {0, 0, "tick"}},
		Expected{"tick", "tock", ""}},
}

func (tc *tc) SetUpQueue() *eventQueue {
	queue := NewEventQueue()
	for _, e := range tc.events {
		queue.Queue(e.mk())
	}
	return queue
}

func (tc *tc) CheckEvents(queue *eventQueue, tick int, t *testing.T) {
	actual := make(map[string]bool)
	for {
		event := queue.GetTriggeredEvent()
		if event == nil {
			break
		}
		if actual[event.Name] == true {
			t.Errorf("Test case \"%s\" failed on tick %d", tc.name, tick)
			t.Errorf("event \"%s\" happened multiple times", event.Name)
		}
		actual[event.Name] = true
	}

	expected := make(map[string]bool)
	for _, e := range strings.Fields(tc.expected[tick]) {
		expected[e] = true
		if !actual[e] {
			t.Errorf("Test case \"%s\" failed on tick %d", tc.name, tick)
			t.Errorf("event \"%s\" is expected but didn't happen", e)
		}
	}

	for e := range actual {
		if !expected[e] {
			t.Errorf("Test case \"%s\" failed on tick %d", tc.name, tick)
			t.Errorf("event \"%s\" happened but is not expected", e)
		}
	}
}

func TestEventQueue(t *testing.T) {
	for _, tc := range testCases {
		queue := tc.SetUpQueue()

		for tick := range tc.expected {
			tc.CheckEvents(queue, tick, t)
			queue.Tick(1)
		}
	}
}

var testCases_remove = []tc{
	{"remove single",
		Events{{0, 1, "tick"}},
		Expected{"tick", "", ""}},
	{"remove first only",
		Events{{2, 0, "tick"}, {3, 0, "tock"}},
		Expected{"", "", "", "tock"}},
	{"remove unexistent",
		Events{{1, 0, "tock"}},
		Expected{"", "tock", "", ""}},
}

func TestRemove(t *testing.T) {
	for _, tc := range testCases_remove {
		scheduler := tc.SetUpQueue()

		for tick := range tc.expected {
			tc.CheckEvents(scheduler, tick, t)
			scheduler.Tick(1)
			if tick == 0 {
				scheduler.Remove("tick")
			}
		}
	}
}

var testCases_add = []tc{
	{"add one",
		Events{},
		Expected{"", "tick", "", "tick", ""}},
	{"add one more",
		Events{{2, 1, "tock"}},
		Expected{"", "tick", "tock", "tick tock"}},
	{"same name",
		Events{{2, 0, "tick"}},
		Expected{"", "", "tick", "", ""}},
}

func TestAdd(t *testing.T) {
	for _, tc := range testCases_add {
		scheduler := tc.SetUpQueue()

		for tick := range tc.expected {
			tc.CheckEvents(scheduler, tick, t)
			scheduler.Tick(1)
			if tick == 0 {
				scheduler.Add(&Event{Name: "tick", Delay: 0, Repeat: 2})
			}
		}
	}
}

var testCases_overrun_plus_4 = []tc{
	{"every tick",
		Events{{0, 1, "tick"}},
		Expected{"tick", "tick", "tick", "tick"}},
	{"every third",
		Events{{0, 3, "tick"}},
		Expected{"tick", "tick", "tick", "", "", "tick"}},
	{"collapse",
		Events{{0, 2, "tick"}, {3, 2, "tock"}},
		Expected{"tick", "tick tock", "tick", "tock"}},
	{"trigger once",
		Events{{1, 3, "tick"}},
		Expected{"", "tick", "", "tick", ""}},
}

func TestOverrun(t *testing.T) {
	for _, tc := range testCases_overrun_plus_4 {
		scheduler := tc.SetUpQueue()

		n := 0
		for tick := range tc.expected {
			tc.CheckEvents(scheduler, tick, t)
			scheduler.Tick(1)
			n++
			if tick == 0 {
				scheduler.Tick(4)
				n += 4
			}
		}
	}
}

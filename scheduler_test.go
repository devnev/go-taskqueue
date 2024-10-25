package taskqueue

import (
	"testing"
	"time"

	"github.com/shoenig/test/must"
)

func TestStartsTask(t *testing.T) {
	start, done := Start(FromMapUnchecked(map[string]Task{"name": {}}))

	select {
	case name := <-start:
		must.EqOp(t, name, "name")
	case <-time.After(time.Second):
		t.Fatal("start channel did not emit")
	}

	select {
	case done <- "name":
	case <-time.After(time.Second):
		t.Fatal("done channel did not accept")
	}

	select {
	case _, ok := <-start:
		must.False(t, ok, must.Sprint("start channel must be closed"))
	case <-time.After(time.Second):
		t.Fatal("start channel did not emit")
	}
}

func TestStartsTaskInOrder(t *testing.T) {
	start, done := Start(FromMapUnchecked(map[string]Task{
		"task1": {},
		"task2": {Deps: []string{"task1"}},
	}))

	select {
	case name := <-start:
		must.EqOp(t, name, "task1")
	case <-time.After(time.Second):
		t.Fatal("start channel did not emit")
	}

	select {
	case done <- "task1":
	case <-time.After(time.Second):
		t.Fatal("done channel did not accept")
	}

	select {
	case name := <-start:
		must.EqOp(t, name, "task2")
	case <-time.After(time.Second):
		t.Fatal("start channel did not emit")
	}

	select {
	case done <- "task2":
	case <-time.After(time.Second):
		t.Fatal("done channel did not accept")
	}

	select {
	case _, ok := <-start:
		must.False(t, ok, must.Sprint("start channel must be closed"))
	case <-time.After(time.Second):
		t.Fatal("start channel did not emit")
	}
}

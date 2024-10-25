// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package taskqueue

import (
	"errors"
	"fmt"
)

var ErrDuplicateName = errors.New("duplicate task name")
var ErrUnknownDep = errors.New("unknown dependency name")
var ErrNameMismatch = errors.New("task name mismatch")
var ErrEmptyName = errors.New("task name is empty")

type Task struct {
	Name  string
	Deps  []string
	Group string
}

type TaskSet struct {
	tasks map[string]Task
}

func FromMapUnchecked(tasks map[string]Task) TaskSet {
	return TaskSet{tasks: tasks}
}

func FromMap(tasks map[string]Task) (TaskSet, error) {
	for name, task := range tasks {
		if name == "" {
			return TaskSet{}, fmt.Errorf("%w", ErrEmptyName)
		}
		if task.Name != "" && task.Name != name {
			return TaskSet{}, fmt.Errorf("%w: %q != %q", ErrNameMismatch, task.Name, name)
		}
	}

	for name, task := range tasks {
		for _, dep := range task.Deps {
			if _, exists := tasks[dep]; !exists {
				return TaskSet{}, fmt.Errorf("%w %q on task %q", ErrUnknownDep, dep, name)
			}
		}
	}

	return TaskSet{tasks: tasks}, nil
}

func FromList(tasks []Task) (TaskSet, error) {
	set := make(map[string]Task, len(tasks))

	for _, task := range tasks {
		if task.Name == "" {
			return TaskSet{}, fmt.Errorf("%w", ErrEmptyName)
		}
		if _, exists := set[task.Name]; exists {
			return TaskSet{}, fmt.Errorf("%w %q", ErrDuplicateName, task.Name)
		}
		set[task.Name] = task
	}

	for name, task := range set {
		for _, dep := range task.Deps {
			if _, exists := set[dep]; !exists {
				return TaskSet{}, fmt.Errorf("%w %q on task %q", ErrUnknownDep, dep, name)
			}
		}
	}

	return TaskSet{tasks: set}, nil
}

func FromMapper[T any](items []T, namer func(T) string, deps func(T) []string, grouper func(T) string) (TaskSet, error) {
	set := make(map[string]Task, len(items))

	for _, item := range items {
		name := namer(item)
		if name == "" {
			return TaskSet{}, fmt.Errorf("%w", ErrEmptyName)
		}
		if _, exists := set[name]; exists {
			return TaskSet{}, fmt.Errorf("%w %q", ErrDuplicateName, name)
		}
		t := Task{Name: name}
		if deps != nil {
			t.Deps = deps(item)
		}
		if grouper != nil {
			t.Group = grouper(item)
		}
	}

	for name, task := range set {
		for _, dep := range task.Deps {
			if _, exists := set[dep]; !exists {
				return TaskSet{}, fmt.Errorf("%w %q on task %q", ErrUnknownDep, dep, name)
			}
		}
	}

	return TaskSet{tasks: set}, nil
}

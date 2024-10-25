// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package taskqueue

import (
	"cmp"
)

// Starts a scheduler for the tasks in the task set.
//
// Names of tasks ready to be executed are sent to the returned start channel.
// On completion of a task, its name must be sent to the done channel.
//
// Tasks will only be scheduled concurrently with other tasks in the same group.
// This also applies to the default group indicated by an empty group name.
//
// The scheduler closes the start channel on exit. It will exit on completion of
// all tasks, or if no further tasks are schedulable (i.e. there is a deadlock),
// or if the done channel is closed.
func Start(tasks TaskSet) (start <-chan string, done chan<- string) {
	startEvents := make(chan string, len(tasks.tasks))
	doneEvents := make(chan string, len(tasks.tasks))

	go func() {
		defer close(startEvents)

		var currentGroup string
		var groupTasksStarted int
		readyTasks := map[string]struct{}{}
		groupReadyCount := map[string]int{}
		startedTasks := map[string]struct{}{}
		completedTasks := map[string]struct{}{}
		peeked := make(chan string, 1)
		for {

			// Handle any pending completion events to give a better idea of the current state of work
			for {
				var completed string
				var doneOpen bool
				select {
				case completed = <-peeked:
				case completed, doneOpen = <-doneEvents:
					if !doneOpen {
						// no more workers, no point in scheduling anything
						return
					}
				default:
				}
				if completed == "" {
					break
				}
				// Mark the received task as completed
				completedTasks[completed] = struct{}{}
				if _, started := startedTasks[completed]; !started {
					// If we receive completion for a task that wasn't even started, accept it and move on
					continue
				}
				delete(startedTasks, completed)
				// Implicit that if a task completes, it is in the current group, as otherwise it wouldn't have been started
				groupTasksStarted -= 1
			}

			// Scan for newly ready tasks
			for name, task := range tasks.tasks {
				if mapHas(completedTasks, name) || mapHas(readyTasks, name) || mapHas(startedTasks, name) {
					continue
				}

				var blocked bool
				for _, dep := range task.Deps {
					if _, exists := tasks.tasks[dep]; !exists {
						continue
					}
					if _, done := completedTasks[dep]; !done {
						blocked = true
						break
					}
				}
				if blocked {
					continue
				}

				readyTasks[name] = struct{}{}
				groupReadyCount[task.Group] += 1
			}

			// If there's no started work on the current group, switch to the group with the maximum ready items
			if groupTasksStarted == 0 {
				currentGroup, _ = maxByValue(groupReadyCount)
			}

			// Try to start an item in the current group
			var nextTask string
			for name := range readyTasks {
				if tasks.tasks[name].Group == currentGroup {
					nextTask = name
					break
				}
			}
			if nextTask != "" {
				select {
				case completed, doneOpen := <-doneEvents:
					if !doneOpen {
						// No more workers, no point in scheduling anything
						return
					}
					peeked <- completed

				case startEvents <- nextTask:
					startedTasks[nextTask] = struct{}{}
					delete(readyTasks, nextTask)
					groupReadyCount[currentGroup] -= 1
					groupTasksStarted += 1
				}

				continue
			}

			// We can't start anything, but there's something running, wait for it to finish
			if len(startedTasks) > 0 {
				completed, doneOpen := <-doneEvents
				if !doneOpen {
					// No more workers, no point in scheduling anything
					return
				}
				peeked <- completed
				continue
			}

			// We're either done or deadlocked
			return
		}
	}()

	return startEvents, doneEvents
}

func mapHas[K comparable, V any, M ~map[K]V](m M, k K) bool {
	_, ok := m[k]
	return ok
}

func maxByValue[K comparable, V cmp.Ordered, M ~map[K]V](m M) (K, V) {
	var minK K
	var minV V

	for k, v := range m {
		minK = k
		minV = v
		break
	}

	for k, v := range m {
		if v > minV {
			minK = k
			minV = v
		}
	}

	return minK, minV
}

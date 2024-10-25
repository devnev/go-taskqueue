// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package taskqueue

import (
	"context"
)

// For use with conc.Pool
func RunOnPool(work <-chan string, completed chan<- string, pool interface {
	Go(func())
	Wait()
}, handler func(string)) {
	defer close(completed)
	for name := range work {
		pool.Go(func() {
			defer func() {
				completed <- name
			}()
			handler(name)
		})
	}
	pool.Wait()
}

// For use with conc.ErrorPool
func RunOnErrorPool(work <-chan string, completed chan<- string, pool interface {
	Go(func() error)
	Wait() error
}, handler func(string) error) error {
	defer close(completed)
	for name := range work {
		pool.Go(func() error {
			defer func() {
				completed <- name
			}()
			return handler(name)
		})
	}
	return pool.Wait()
}

// For use with conc.ContextPool
func RunOnContextPool(work <-chan string, completed chan<- string, pool interface {
	Go(func(context.Context) error)
	Wait() error
}, handler func(context.Context, string) error) error {
	defer close(completed)
	for name := range work {
		pool.Go(func(ctx context.Context) error {
			defer func() {
				completed <- name
			}()
			return handler(ctx, name)
		})
	}
	return pool.Wait()
}

// For use with conc.ResultPool
func RunOnResultPool[T any](work <-chan string, completed chan<- string, pool interface {
	Go(func() T)
	Wait() []T
}, handler func(string) T) []T {
	defer close(completed)
	for name := range work {
		pool.Go(func() T {
			defer func() {
				completed <- name
			}()
			return handler(name)
		})
	}
	return pool.Wait()
}

// For use with conc.ResultErrorPool
func RunOnResultErrorPool[T any](work <-chan string, completed chan<- string, pool interface {
	Go(func() (T, error))
	Wait() ([]T, error)
}, handler func(string) (T, error)) ([]T, error) {
	defer close(completed)
	for name := range work {
		pool.Go(func() (T, error) {
			defer func() {
				completed <- name
			}()
			return handler(name)
		})
	}
	return pool.Wait()
}

// For use with conc.ResultContextPool
func RunOnResultContextPool[T any](work <-chan string, completed chan<- string, pool interface {
	Go(func(context.Context) (T, error))
	Wait() ([]T, error)
}, handler func(context.Context, string) (T, error)) ([]T, error) {
	defer close(completed)
	for name := range work {
		pool.Go(func(ctx context.Context) (T, error) {
			defer func() {
				completed <- name
			}()
			return handler(ctx, name)
		})
	}
	return pool.Wait()
}

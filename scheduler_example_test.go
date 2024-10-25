package taskqueue_test

import (
	"fmt"

	"github.com/devnev/go-taskqueue"
	"github.com/sourcegraph/conc/pool"
)

func Example() {
	tasks, err := taskqueue.FromList([]taskqueue.Task{{
		Name: "init",
	}, {
		Name: "partA",
		Deps: []string{"init"},
	}, {
		Name: "partB",
		Deps: []string{"partB"},
	}, {
		Name: "finalize",
		Deps: []string{"partA", "partB"},
	}})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	work, done := taskqueue.Start(tasks)
	taskqueue.RunOnPool(work, done, pool.New(), func(name string) {
		fmt.Printf("Running %s\n", name)
	})
}

package errors

import (
	"sync"
)

// Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// This differs from errgroup.Group in that this doesn't cancel the group when subtask
// returns an error. Instead this accumulates errors from all subtasks in a List.
type Group struct {
	mut  sync.Mutex
	list List
	wg   sync.WaitGroup
}

// Go calls the given function in a new goroutine.
func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.mut.Lock()
			g.list.Add(err)
			g.mut.Unlock()
		}
	}()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns a List of non-nil errors returned from all function calls.
func (g *Group) Wait() error {
	g.wg.Wait()
	if len(g.list) == 0 {
		return nil
	}
	return g.list
}

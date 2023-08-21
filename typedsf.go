package typedsf

import (
	"fmt"

	"golang.org/x/sync/singleflight"
)

// Result is the same as singleflight.Result, but with type information
type Result[T any] struct {
	Val    T
	Err    error
	Shared bool
}

// Group is the same as singleflight.Group, but with type information
type Group[T any] struct {
	sf *singleflight.Group
}

func (g *Group[T]) setSingleflight() {
	if g.sf == nil {
		g.sf = new(singleflight.Group)
	}
}

// Do is the same as singleflight.Group, but with type information. If for some
// reason an incorrect type is returned by `fn` then an error will be returned
// to all callers.
func (g *Group[T]) Do(key string, fn func() (T, error)) (T, error, bool) {
	g.setSingleflight()
	val, err, shared := g.sf.Do(key, func() (interface{}, error) {
		return fn()
	})
	var typedVal T // create default
	if tempVal, ok := val.(T); !ok {
		return typedVal, fmt.Errorf("typedsf: expected %T but got %T", typedVal, val), shared
	} else {
		return tempVal, err, shared
	}
}

// DoChan is the same as singleflight.Group, but with type information. If for some
// reason an incorrect type is returned by `fn` then an error will be returned
// to all callers via the channel.
//
// The returned channel is not closed.
func (g *Group[T]) DoChan(key string, fn func() (T, error)) <-chan Result[T] {
	g.setSingleflight()
	ch := make(chan Result[T], 1)
	go func() {
		ch2 := g.sf.DoChan(key, func() (interface{}, error) {
			return fn()
		})
		res := <-ch2
		var typedVal T
		if res.Val != nil {
			if tempVal, ok := res.Val.(T); !ok {
				ch <- Result[T]{
					Val:    typedVal, // default
					Err:    fmt.Errorf("typedsf: expected %T but got %T", typedVal, res.Val),
					Shared: res.Shared,
				}
				return
			} else {
				typedVal = tempVal
			}
		}
		ch <- Result[T]{
			Val:    typedVal,
			Err:    res.Err,
			Shared: res.Shared,
		}
	}()
	return ch
}

// Forget is the same as singleflight.Group
func (g *Group[T]) Forget(key string) {
	g.setSingleflight()
	g.sf.Forget(key)
}

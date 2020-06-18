package cascade

import (
	"reflect"
	"strings"
	"sync"
)

func removePointerAsterisk(s string) string {
	return strings.Trim(s, "*")
}

func isReference(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

// TODO add all primitive types
func isPrimitive(str string) bool {
	switch str {
	case "int":
		return true
	default:
		return false
	}
}

// waitDone is wrapper on ctx.Done channel
// which will wait until cancel() will be invoked
func (c *Cascade) waitDone() {
	for {
		select {
		case <-c.ctx.Done():
			return
		}
	}
}

func merge(in []*Result) <-chan *Result {
	var wg sync.WaitGroup
	out := make(chan *Result)

	output := func(r *Result) {
		for range r.ErrCh {
			out <- r
		}
		wg.Done()
	}

	wg.Add(len(in))

	for _, c := range in {
		go output(c)
	}

	// TODO close on stop
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

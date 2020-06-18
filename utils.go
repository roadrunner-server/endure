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

func merge(in []*result) <-chan *Result {
	var wg sync.WaitGroup
	out := make(chan *Result)

	output := func(r *result) {
		for k := range r.errCh {
			if k == nil {
				continue
			}
			out <- &Result{
				Err:      k,
				VertexID: r.vertexId,
			}
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

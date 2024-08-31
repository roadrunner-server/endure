package endure

import (
	"go.uber.org/zap"
)

// poll is used to poll the errors from the vertex
func (e *Endure) poll(r *result) {
	go func(res *result) {
		for err := range res.errCh {
			if err == nil {
				continue
			}
			// log error message
			e.log.Error("plugin returned an error from the 'Serve' method", zap.Error(err), zap.String("plugin", res.vertexID))
			// set the error
			res.err = err
			// send handleErrorCh signal
			e.handleErrorCh <- res
		}
	}(r)
}

func (e *Endure) startMainThread() {
	// main thread used to handle errors from vertices
	go func() {
		for res := range e.handleErrorCh {
			e.log.Debug("processing error in the main thread", zap.String("id", res.vertexID))
			e.userResultsCh <- &Result{
				Error:    res.err,
				VertexID: res.vertexID,
			}
		}
	}()
}

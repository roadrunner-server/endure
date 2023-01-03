package endure

import (
	"golang.org/x/exp/slog"
)

// poll is used to poll the errors from the vertex
// and exit from it
func (e *Endure) poll(r *result) {
	rr := r
	go func(res *result) {
		for {
			select {
			// error
			case err := <-res.errCh:
				if err != nil {
					// log error message
					e.log.Error("plugin returned an error from the Serve", err, slog.String("id", res.vertexID))

					// set the error
					res.err = err

					// send handleErrorCh signal
					e.handleErrorCh <- res
				}
			// exit from the goroutine
			case <-res.signal:
				e.log.Info("shutdown signal received", slog.String("id", res.vertexID))
				return
			}
		}
	}(rr)
}

func (e *Endure) startMainThread() {
	/*
		Main thread is the main Endure unit of work
		It used to handle errors from vertices, notify user about result, re-calculating graph according to failed vertices and sending internal_stop signals
	*/
	go func() {
		for res := range e.handleErrorCh {
			e.log.Debug("processing error in the main thread", slog.String("id", res.vertexID))
			e.userResultsCh <- &Result{
				Error:    res.err,
				VertexID: res.vertexID,
			}
		}
	}()
}

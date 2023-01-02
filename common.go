package endure

func (e *Endure) sendResultToUser(res *result) {
	e.userResultsCh <- &Result{
		Error:    res.err,
		VertexID: res.vertexID,
	}
}

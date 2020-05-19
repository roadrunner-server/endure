package cascade

import (
	"github.com/spiral/cascade/logger"
)

type Demo struct {
	logger logger.Logger
	d2     Demo2
}

func (d *Demo) Init(logger logger.Logger, d2 Demo2) error {
	d.logger = logger
	d.d2 = d2

	return Disabled
}

func (d *Demo) RPC() interface{} {
	return nil
}

// ?????
func (d *Demo) Serve(health chan error) error {
	// 1: we don't when service alive
	// 2: we don't know when service stopped
	return nil
}

// ???
func (d *Demo) Stop() {
	// return ?
}

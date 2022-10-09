package plugin6

import (
	"fmt"
	"net/http"

	"github.com/roadrunner-server/errors"
)

type Plugin6 struct {
}

func (p6 *Plugin6) Init() error {
	go func() {
		_ = http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:gosec
			_, _ = fmt.Fprint(w, "hello")
		}))
	}()

	return errors.E(errors.Op("plugin6 init"), errors.Disabled)
}

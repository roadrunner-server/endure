package plugin9

import "fmt"

type Plugin9 struct {
}

// No deps
func (s *Plugin9) Init() error {
	return nil
}

func (s *Plugin9) Name() string {
	return "Plugin9"
}

func (s *Plugin9) Boom() {
	fmt.Println("Boom from plugin9")
}

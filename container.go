package endure

// InitMethodName is the function fn for the reflection
const InitMethodName = "Init"

// ServeMethodName is the function fn for the Serve
const ServeMethodName = "Serve"

// StopMethodName is the function fn for the reflection to Stop the service
const StopMethodName = "Stop"

// Result is the information which endure send to the user
type Result struct {
	Error    error
	VertexID string
}

type notify struct{}

type result struct {
	// error channel from vertex
	errCh chan error
	// error from the channel
	err error
	// unique vertex id
	vertexID string
	// notify used to signal vertex about event
	signal chan notify
}

type (
	// this is the main service interface with should implement every plugin
	Service interface {
		// Serve
		Serve() chan error
		// Stop
		Stop() error
	}

	// Name of the service
	Named interface {
		Name() string
	}

	// internal container interface
	Container interface {
		Serve() (<-chan *Result, error)
		Stop() error
		Register(service interface{}) error
		Init() error
	}

	// Provider declares the ability to provide service edges of declared types.
	Provider interface {
		Provides() []interface{}
	}

	// Collector declares the ability to accept the plugins which match the provided method signature.
	Collector interface {
		Collects() []interface{}
	}
)

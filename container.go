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
	// Service interface can be implemented by the plugin to use Start-Stop functionality
	Service interface {
		// Serve starts the plugin
		Serve() chan error
		// Stop stops the plugin
		Stop() error
	}

	// Name of the service
	Named interface {
		// Name return user friendly name of the plugin
		Name() string
	}

	// internal container interface
	Container interface {
		// Serve used to Start the plugin in topological order
		Serve() (<-chan *Result, error)
		// Stop stops the plugins in rev-topological order
		Stop() error
		// Register registers one plugin in container
		Register(service interface{}) error
		// RegisterAll register set of comma separated plugins in container
		RegisterAll(service ...interface{}) error
		// Init initializes all plugins (calling Init function), calculate vertices, invoke Collects and Provided functions if exist
		Init() error
	}

	// Provider declares the ability to provide service edges of declared types.
	Provider interface {
		// Provides function return set of functions which provided dependencies to other plugins
		Provides() []interface{}
	}

	// Collector declares the ability to accept the plugins which match the provided method signature.
	Collector interface {
		// Collects search for the structures or (and) interfaces in the arguments and provides it for the plugin
		Collects() []interface{}
	}
)

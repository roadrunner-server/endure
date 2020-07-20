package cascade

// InitMethodName is the function name for the reflection
const InitMethodName = "Init"

// ConfigureMethodName
const ConfigureMethodName = "Configure"

// CloseMethodName
const CloseMethodName = "Close"

// ServeMethodName
const ServeMethodName = "Serve"

// Stop is the function name for the reflection to Stop the service
const StopMethodName = "Stop"

type Result struct {
	Error    Error
	VertexID string
}

type result struct {
	// error channel from vertex
	errCh chan error
	// error from the channel
	err error
	// unique vertex id
	vertexId string
	// signal to the vertex goroutine to exit
	exit chan struct{}
	// internal exit, used to notify main thread to release resources
	internalExit bool
}

type (
	// used to gracefully stop and configure the plugins
	Graceful interface {
		// Configure is used when we need to make preparation and wait for all services till Serve
		Configure() error
		// Close frees resources allocated by the service
		Close() error
	}
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
	container interface {
		Serve() (error, <-chan *Result)
		Stop() error
		restart() error
		Register(service interface{}) error
		Init() error
	}

	// Provider declares the ability to provide service edges of declared types.
	Provider interface {
		Provides() []interface{}
	}

	// Depender declares the ability to accept the plugins which match the provided method signature.
	Depender interface {
		Depends() []interface{}
	}
)

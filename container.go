package endure

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

const (
	// InitMethodName is the function fn for the reflection
	InitMethodName = "Init"
	// ServeMethodName is the function fn for the Serve
	ServeMethodName = "Serve"
	// StopMethodName is the function fn for the reflection to Stop the service
	StopMethodName = "Stop"
)

// Result is the information which endure sends to the user
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
		Stop(context.Context) error
	}

	// Named -> Name of the service
	Named interface {
		// Name return user friendly name of the plugin
		Name() string
	}

	// Container - Internal container interface
	Container interface {
		// Serve used to Start the plugin in topological order
		Serve() (<-chan *Result, error)
		// Stop stops the plugins in rev-topological order
		Stop() error
		// Register registers one plugin in container
		Register(service any) error
		// Plugins method is responsible for returning an all registered plugins
		Plugins() string
		// RegisterAll register set of comma separated plugins in container
		RegisterAll(service ...any) error
		// Init initializes all plugins (calling Init function), calculate vertices, invoke Collects and Provided functions if exist
		Init() error
	}

	// Provider declares the ability to provide service edges of declared types.
	Provider interface {
		// Provides function return set of functions which provided dependencies to other plugins
		Provides() []*dep.Out
	}

	// Weighted is optional to implement, but when implemented the return value added during the topological sort
	Weighted interface {
		Weight() uint
	}

	// Collector declares the ability to accept the plugins which match the provided method signature.
	Collector interface {
		// Collects search for the plugins which implements given interfaces in the args
		Collects() []*dep.In
	}
)

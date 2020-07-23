# Cascade
<p align="center">
	<a href="https://pkg.go.dev/github.com/spiral/cascade?tab=doc"><img src="https://godoc.org/github.com/spiral/cascade?status.svg"></a>
	<a href="https://github.com/spiral/cascade/actions"><img src="https://github.com/spiral/cascade/workflows/CI/badge.svg" alt=""></a>
	<a href="https://goreportcard.com/report/github.com/spiral/cascade"><img src="https://goreportcard.com/badge/github.com/spiral/cascade"></a>
	<a href="https://codecov.io/gh/spiral/cascade/"><img src="https://codecov.io/gh/spiral/cascade/branch/master/graph/badge.svg"></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
</p>

Cascade is an open-source (MIT licensed) plugin container.

<h2>Features</h2>

- Production ready
- Supports structs and interfaces (see examples)
- Use graph to topologically sort, run and stop dependent plugins
- Algorithm used: graph and double-linked list
- Support easy to add Middleware plugins
- Error reporting
- Automatically restart failing vertices


<h2>Installation</h2>  

```go
go get -u github.com/spiral/cascade
```  


<h2>God damn WHY?</h2>  

Imagine you have an application in which you want to implement plugin system. These plugins can depend on each other (via interfaces or directly).
For example, we have 3 plugins: HTTP (to communicate with world), DB (to save the world) and logger (to see the progress).  
In this case, we can't start HTTP before we start all other parts. Also, we need to have logger first. So, the order will be the following:  
1. Initialize the logger
2. Initialize the DB
3. Initialize the HTTP  
Ok, next we need to start it, and in case of error - restart or stop in reverse order. All you need to do in Cascade is to pass HTTP, DB and logger structs to cascade and implement cascade interface. That's it. Cascade will take care of restarting failing vertices (structs, HTTP for example) with exponential backoff mechanism.  

<h2>Cascade main interface</h2>  

```go
package sample

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

	// Provider declares the ability to provide service edges of declared types.
	Provider interface {
		Provides() []interface{}
	}

	// Depender declares the ability to accept the plugins which match the provided method signature.
	Depender interface {
		Depends() []interface{}
	}
)  
```
Order is the following:
1. `Init() error` - mandatory to implement. In your structure (which you pass to Cascade), you should have this method as receiver. It can accept as parameter any passed to cascade structure (see sample) or interface (with limitations).  
2. `Graceful` - optional to implement. Used to configure a vertex before invoking `Serve` method. Has the `Confugure` method which will be invoked after `Init` and `Close` which will be invoked after `Stop` to free some resources for example.
3. `Service` - mandatory to implement. Has 2 main methods - `Serve` which should return initialized golang channel with errors, and `Stop` to stop the shutdown the Cascade.
4. `Provider` - optional to implement. Used to provide some dependency if you need to extend your struct.
5. `Depender` - optional to implement. Used to mark structure (vertex) as some struct dependency. It can accept interfaces which implement caller.
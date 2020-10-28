# Endure [currently in beta]
<p align="center">
	<a href="https://pkg.go.dev/github.com/spiral/Endure?tab=doc"><img src="https://godoc.org/github.com/spiral/Endure?status.svg"></a>
	<a href="https://github.com/spiral/Endure/actions"><img src="https://github.com/spiral/Endure/workflows/CI/badge.svg" alt=""></a>
	<a href="https://codecov.io/gh/spiral/endure"><img src="https://codecov.io/gh/spiral/endure/branch/master/graph/badge.svg?token=itNaiZ6ALN"/></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
	<a href="https://lgtm.com/projects/g/spiral/endure/alerts/"><img src="https://img.shields.io/lgtm/alerts/g/spiral/endure.svg?logo=lgtm&logoWidth=18"></a>
</p>

Endure is an open-source (MIT licensed) plugin container with IoC and self-healing.

<h2>Features</h2>

- Supports structs and interfaces (see examples)
- Use graph to topologically sort, run, stop and restart dependent plugins
- Algorithms used: graph and double-linked list
- Support easy to add Middleware plugins
- Error reporting
- Automatically restart failing vertices


<h2>Installation</h2>  

```go
go get -u github.com/spiral/endure
```  


<h2>Why?</h2>  

Imagine you have an application in which you want to implement plugin system. These plugins can depend on each other (via interfaces or directly).
For example, we have 3 plugins: HTTP (to communicate with world), DB (to save the world) and logger (to see the progress).  
In this particular case, we can't start HTTP before we start all other parts. Also, we have to initialize logger first, because all parts of our system need logger. All you need to do in `Endure` is to pass `HTTP`, `DB` and `Logger` structs to the `Endure` and implement `Endure` interface. So, the dependency graph will be the following:
<p align="left">
  <img src="https://github.com/spiral/endure/blob/master/images/graph.png" width="300" height="250" />
</p>

Next we need to start serving all the parts:
```go
errCh, err := container.Serve()
```
`errCh` is the channel with errors from the all `Vertices`. You can identify vertex by `vertexID` which is presented in `errCh` struct.
Then just process the events from the `errCh`:
```go
for {
	select {
	case e := <-errCh:
		println(e.Error.Err.Error()) // just print the error, but actually error processing could be there
		er := container.Stop()
		if er != nil {
		    panic(er)
		}
		return
	}
}
```
Also `Endure` will take care of the restart failing vertices (HTTP, DB, Logger in example) with exponential backoff mechanism.   
The start will proceed in topological order (`Logger` -> `DB` -> `HTTP`), and the stop in reverse-topological order automatically.


<h2>Endure main interface</h2>  

```go
package sample

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

	// Provider declares the ability to provide service edges of declared types.
	Provider interface {
		Provides() []interface{}
	}

	// Collector declares the ability to accept the plugins which match the provided method signature.
	Collector interface {
		Collects() []interface{}
	}
)  
```
Order is the following:
1. `Init() error` - is mandatory to implement. In your structure (which you pass to Endure), you should have this method as a receiver. It can accept as parameter any passed to the `Endure` structure (see samples) or interface (with limitations).  
3. `Service` - is optional to implement. It has 2 main methods - `Serve` which should run the plugin and return initialized golang channel with errors, and `Stop` to shut down the plugin.
4. `Provider` - is optional to implement. It is used to provide some dependency if you need to extend your struct.
5. `Collector` - is optional to implement. It is used to mark a structure (vertex) as some struct dependency. It can accept interfaces which implement the caller.
6. `Named` - is optional to implement. This is a special kind of interface which provides the name of the struct (plugin, vertex) to the caller. Is useful in logger (for example) to know user-friendly plugin name.

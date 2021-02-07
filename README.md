# Endure [currently in beta]

<p align="center">
 <a href="https://pkg.go.dev/github.com/spiral/Endure?tab=doc"><img src="https://godoc.org/github.com/spiral/Endure?status.svg"></a>
 <a href="https://github.com/spiral/Endure/actions"><img src="https://github.com/spiral/Endure/workflows/Linux/badge.svg" alt=""></a>
 <a href="https://github.com/spiral/Endure/actions"><img src="https://github.com/spiral/Endure/workflows/macOS/badge.svg" alt=""></a>
 <a href="https://github.com/spiral/Endure/actions"><img src="https://github.com/spiral/Endure/workflows/Windows/badge.svg" alt=""></a>
 <a href="https://github.com/spiral/Endure/actions"><img src="https://github.com/spiral/Endure/workflows/Linters/badge.svg" alt=""></a>
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

### Why?

Imagine you have an application in which you want to implement plugin system. These plugins can depend on each other (
via interfaces or directly). For example, we have 3 plugins: HTTP (to communicate with world), DB (to save the world)
and logger (to see the progress).  
In this particular case, we can't start HTTP before we start all other parts. Also, we have to initialize logger first,
because all parts of our system need logger. All you need to do in `Endure` is to pass `HTTP`, `DB` and `Logger` structs
to the `Endure` and implement `Endure` interface. So, the dependency graph will be the following:

<p align="left">
  <img src="https://github.com/spiral/endure/blob/master/images/graph.png" width="300" height="250" />
</p>

=======  
First step is to initialize the `endure` container:

```go
container, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.Visualize(endure.StdOut, ""))
```

Let's take a look at the `endure.NewContainer()`:

1. First arg here is the external logger. If you want to use your own logger, you can pass it as the first argument.
2. Next arguments are optional and can be set using `Options`. For example `endure.Visualize(endure.StdOut, "")` will
   show you dot-compatible graph in the console. TODO: all options section.  
   Then we need to pass our structures as references to the `RegisterAll` or `Register` function.

```go
err = container.RegisterAll(
		&httpPlugin{},
		&DBPlugin{},
		&LoggerPlugin{},
	)
if err != nil {
    panic(err)
}
```

The order of plugins in the `RegisterAll` plugin does no matter.  
Next we need to initialize and run our container:

```go
err = container.Init()
if err != nil {
    panic(err)
}
errCh, err := container.Serve()
if err != nil {
    panic(err)
}
```

`errCh` is the channel with errors from the all `Vertices`. You can identify vertex by `vertexID` which is presented
in `errCh` struct. Then just process the events from the `errCh`:

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

Also `Endure` will take care of the restart failing vertices (HTTP, DB, Logger in example) with exponential backoff
mechanism.  
The start will proceed in topological order (`Logger` -> `DB` -> `HTTP`), and the stop in reverse-topological order
automatically.

<h2>Endure main interface</h2>

```go
package sample

type (
	// This is the main Endure service interface which may be implemented to Start (Serve) and Stop plugin (OPTIONAL)
	Service interface {
		// Serve
		Serve() chan error
		// Stop
		Stop() error
	}

	// Name of the service (OPTIONAL)
	Named interface {
		Name() string
	}

	// Provider declares the ability to provide dependencies to other plugins (OPTIONAL)
	Provider interface {
		Provides() []interface{}
	}

	// Collector declares the ability to accept the plugins which match the provided method signature (OPTIONAL)
	Collector interface {
		Collects() []interface{}
	}
)

// Init is mandatory to implement
type Plugin struct{}

func (p *Plugin) Init( /* deps here */) error {
	return nil
}
```

Order is the following:

1. `Init() error` - is mandatory to implement. In your structure (which you pass to Endure), you should have this method
   as a receiver. It can accept as parameter any passed to the `Endure` structure (see samples) or interface (with
   limitations).
2. `Service` - is optional to implement. It has 2 main methods - `Serve` which should run the plugin and return
   initialized golang channel with errors, and `Stop` to shut down the plugin.
3. `Provider` - is optional to implement. It is used to provide some dependency if you need to extend your struct.
4. `Collector` - is optional to implement. It is used to mark a structure (vertex) as some struct dependency. It can
   accept interfaces which implement the caller.
5. `Named` - is optional to implement. This is a special kind of interface which provides the name of the struct (
   plugin, vertex) to the caller. Is useful in logger (for example) to know user-friendly plugin name.

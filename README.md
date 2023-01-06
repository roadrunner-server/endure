# Endure

<p align="center">
 <a href="https://pkg.go.dev/github.com/roadrunner-server/endure?tab=doc"><img src="https://godoc.org/github.com/roadrunner-server/endure?status.svg"></a>
 <a href="https://github.com/roadrunner-server/endure/actions"><img src="https://github.com/roadrunner-server/endure/workflows/Linux/badge.svg" alt=""></a>
 <a href="https://github.com/roadrunner-server/endure/actions"><img src="https://github.com/roadrunner-server/endure/workflows/macOS/badge.svg" alt=""></a>
 <a href="https://github.com/roadrunner-server/endure/actions"><img src="https://github.com/roadrunner-server/endure/workflows/Windows/badge.svg" alt=""></a>
 <a href="https://github.com/roadrunner-server/endure/actions"><img src="https://github.com/roadrunner-server/endure/workflows/Linters/badge.svg" alt=""></a>
 <a href="https://codecov.io/gh/roadrunner-server/endure"><img src="https://codecov.io/gh/roadrunner-server/endure/branch/master/graph/badge.svg?token=itNaiZ6ALN"/></a>
 <a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
 <a href="https://lgtm.com/projects/g/roadrunner-server/endure/alerts/"><img src="https://img.shields.io/lgtm/alerts/g/roadrunner-server/endure.svg?logo=lgtm&logoWidth=18"></a>
</p>

Endure is an open-source (MIT licensed) plugin container with IoC (Inversion of Control) and self-healing capabilities.

## Features

- Supports interfaces (see examples)
- Uses a graph to topologically sort, run, stop, and restart dependent plugins
- Supports easy addition of Middleware plugins
- Error reporting

## Installation

```go
go get -u github.com/roadrunner-server/endure/v2
```


### Why?

Imagine you have an application in which you want to implement a plugin system. These plugins can depend on each other (via interfaces or directly). For example, we have 3 plugins: HTTP (to communicate with the world), DB (to save the world), and logger (to see the progress). In this particular case, we can't start HTTP before we start all other parts. Also, we have to initialize the logger first, because all parts of our system need the logger. All you need to do in Endure is to pass the `HTTP`, `DB`, and `Logger` structs to Endure and implement the `Endure` interface. So, the dependency graph will be the following:

![Dependency Graph](https://github.com/roadrunner-server/endure/blob/master/images/graph.png)

First, we initialize the `endure` container:

```go
import (
    "golang.org/x/exp/slog"
)

func main() {
    container := endure.New(slog.LevelDebug, endure.Visualize())
}
```

Let's take a look at the `endure.New()` function:

1. The first argument is the standard golang logger log level.
2. The next arguments are optional and can be set using `Options`. For example, `endure.Visualize()` will show you a dot-compatible graph in the console. Then we need to pass our structures as references to the `RegisterAll` or `Register` function.


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

The order of plugins in the `RegisterAll` function does not matter.
Next, we need to initialize and run our container:


```go
err := container.Init()
    if err != nil {
        panic(err)
}
errCh, err := container.Serve()
    if err != nil {
    	panic(err)
}
```


`errCh` is the channel with errors from all `Vertices`. You can identify the vertex by `vertexID`, which is presented in the `errCh` struct. Then just process the events from the `errCh`:

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

The start will proceed in topological order (Logger -> DB -> HTTP), and the stop in reverse-topological order automatically.

### Endure main interface

```go
package sample

import (
   "github.com/roadrunner-server/endure/v2/dep"
)

type (
   // This is the main Endure service interface which may be implemented to Start (Serve) and Stop plugin (OPTIONAL)
   Service interface {
      // Serve
      Serve() chan error
      // Stop with context, if you reach the timeout, endure will force the exit via context deadline
      Stop(context.Context) error
      // Named return plugin's name
      Named() string
   }

   // Provider declares the ability to provide dependencies to other plugins (OPTIONAL)
   Provider interface {
      Provides() []*dep.In
   }

   // Collector declares the ability to accept the plugins which match the provided method signature (OPTIONAL)
   Collector interface {
      Collects() []*dep.Out
   }
)

// Init is mandatory to implement
type Plugin struct{}

func (p *Plugin) Init( /* deps here */) error {
   return nil
}
```

Order is the following:

1. `Init() error` - is mandatory to implement. For your structure (which you pass to `Endure`), you should have this method as the method of the struct (```go func (p *Plugin) Init() error {}```). It can accept as a parameter any passed to the `Endure` structure (see samples) or interface (with limitations).
2. `Service` - is optional to implement. It has 2 methods: `Serve` which should run the plugin and return an initialized golang channel with errors, and `Stop` to shut down the plugin. The `Stop` and `Serve` should not block the execution.
3. `Provider` - is optional to implement. It is used to provide some dependency if you need to extend your struct without deep modification.
4. `Collector` - is optional to implement. It is used to mark a structure (vertex) as some struct dependency. It can accept interfaces that implement a caller.
5. `Named` - is mandatory to implement. This is a special kind of interface that provides the name of the struct (plugin, vertex) to the caller. It is useful in the logger (for example) to know the user-friendly plugin name.

Available options:
1. `Visualize`: Graph visualization option via graphviz. The Graphviz diagram can be shown via stdout.
2. `GracefulShutdownTimeout`: `time.Duration`. How long to wait for a vertex (plugin) to stop.

The fully operational example is located in the `examples` folder.

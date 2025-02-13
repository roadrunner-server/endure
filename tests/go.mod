module github.com/roadrunner-server/endure/v2/tests

go 1.24

toolchain go1.24.0

replace github.com/roadrunner-server/endure/v2 => ../

require (
	github.com/roadrunner-server/endure/v2 v2.0.0-00010101000000-000000000000
	github.com/roadrunner-server/errors v1.4.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

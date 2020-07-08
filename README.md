# cascade
<p align="center">
	<a href="https://pkg.go.dev/github.com/spiral/cascade?tab=doc"><img src="https://godoc.org/github.com/spiral/cascade?status.svg"></a>
	<a href="https://github.com/spiral/cascade/actions"><img src="https://github.com/spiral/cascade/workflows/CI/badge.svg" alt=""></a>
	<a href="https://goreportcard.com/report/github.com/spiral/cascade"><img src="https://goreportcard.com/badge/github.com/spiral/cascade"></a>
	<a href="https://codecov.io/gh/spiral/cascade/"><img src="https://codecov.io/gh/spiral/cascade/branch/master/graph/badge.svg"></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
</p>


Draft:

- Init can accept only 1 implementation of interface. For example, only 1 logger implementation at the time may exist.
But, if in the system exists more than 1 implementation, Init accepts first in order.
- Depends can accept all implementations (like in examples -> http Middleware)
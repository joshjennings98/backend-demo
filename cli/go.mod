module github.com/joshjennings98/backend-demo/cli

replace github.com/joshjennings98/backend-demo/server => ../server

go 1.22.2

require github.com/joshjennings98/backend-demo/server v0.0.0-00010101000000-000000000000

require (
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/maragudk/gomponents v0.20.1 // indirect
	github.com/maragudk/gomponents-htmx v0.4.0 // indirect
	github.com/yuin/goldmark v1.4.13 // indirect
	golang.org/x/net v0.25.0 // indirect
)

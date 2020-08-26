package main

// Bare minimum to work with this module locally
import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver"
	_ "github.com/jaysonsantos/caddy-archiver"
)

func main() {
	caddycmd.Main()
}

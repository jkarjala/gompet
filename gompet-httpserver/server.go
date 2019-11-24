// This file is part of Gompet - Copyright 2019 Jari Karjala - www.jpkware.com
// SPDX-License-Identifier: GPLv3-only

// Simple fasthttp server for testing the HTTP client
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

var listenAddr = flag.String("a", "127.0.0.1:4200", "Address and port to listen")

func main() {
	flag.Parse()
	fmt.Println("Server listening at", *listenAddr)
	if err := fasthttp.ListenAndServe(*listenAddr, requestHandler); err != nil {
		log.Fatalf("error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	switch s := string(ctx.Path()); s {
	case "/":
		ctx.SetContentType("text/plain")
		ctx.Write([]byte("OK"))
	case "/ping":
		ctx.SetContentType("text/plain")
		ctx.Write([]byte("PONG"))
	case "/get", "/put", "/post", "/patch", "/delete", "/echo":
		ctx.SetContentType(string(ctx.Request.Header.ContentType()))
		ctx.Write([]byte(s))
		ctx.Write([]byte(":"))
		ctx.Write(ctx.Request.Body())
	case "/snoop":
		snoop(ctx)
	default:
		parts := strings.Split(s, "/")
		if parts[1] == "status" && len(parts) > 2 {
			code, _ := strconv.Atoi(parts[2])
			ctx.SetStatusCode(code)
		} else {
			ctx.SetStatusCode(404)
		}
	}
}

func snoop(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Request method is %q\n", ctx.Method())
	fmt.Fprintf(ctx, "RequestURI is %q\n", ctx.RequestURI())
	fmt.Fprintf(ctx, "Requested path is %q\n", ctx.Path())
	fmt.Fprintf(ctx, "Host is %q\n", ctx.Host())
	fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())
	fmt.Fprintf(ctx, "Content-Type is %q\n", ctx.Request.Header.ContentType())
	fmt.Fprintf(ctx, "Content-Length is %q\n", ctx.Request.Header.ContentLength())
	fmt.Fprintf(ctx, "User-Agent is %q\n", ctx.UserAgent())
	fmt.Fprintf(ctx, "Connection has been established at %s\n", ctx.ConnTime())
	fmt.Fprintf(ctx, "Request has been started at %s\n", ctx.Time())
	fmt.Fprintf(ctx, "Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
	fmt.Fprintf(ctx, "Your ip is %q\n\n", ctx.RemoteIP())

	fmt.Fprintf(ctx, "Raw request is:\n---CUT---\n%s\n---CUT---", &ctx.Request)

	ctx.SetContentType("text/plain; charset=utf8")
}

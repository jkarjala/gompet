// This file is part of Gompet - Copyright 2019-2020 Jari Karjala - www.jpkware.com
// SPDX-License-Identifier: GPLv3-only

// HTTP Client using the fasthttp library
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jkarjala/gompet"
	"github.com/valyala/fasthttp"
)

// Long options for the testers, short ones used by the main library
var httpAuth = flag.String("auth", "", "HTTP Authorization header")
var httpContentType = flag.String("content-type", "application/json", "HTTP body content type")

func main() {
	log.Println("FastHTTP tester started")
	gompet.Run(clientFactory)
}

type myClient struct {
	config     gompet.ClientConfig
	req        *fasthttp.Request
	res        *fasthttp.Response
	httpClient *fasthttp.Client
}

func clientFactory(config gompet.ClientConfig) (gompet.Client, error) {
	log.Println(config.ID, "fasthttp init")

	var req = fasthttp.AcquireRequest()
	var res = fasthttp.AcquireResponse()
	var httpClient = &fasthttp.Client{}
	var client = myClient{config, req, res, httpClient}
	return &client, nil
}

func (c *myClient) RunCommand(in *gompet.ClientInput) *gompet.ClientResult {
	cmd := in.Cmd
	if c.config.Template != nil {
		cmd = c.config.Template.Expand(in.Args)
	}

	var err error

	var ss = strings.SplitN(cmd, " ", 3)
	if len(ss) < 2 {
		panic(fmt.Sprintf("Invalid command %s, HTTP verb and URL required", cmd))
	}
	var primitive = strings.Trim(ss[0], "\r\t ")
	var url = strings.Trim(ss[1], "\r\t ")
	var body = ""
	if len(ss) == 3 {
		body = strings.Trim(ss[2], "\r\t")
	}

	c.req.Reset()
	c.req.Header.SetMethod(primitive)
	c.req.SetRequestURI(url)

	if body != "" {
		c.req.Header.Set("Content-Type", *httpContentType)
		c.req.SetBody([]byte(body))
	}
	if *httpAuth != "" {
		c.req.Header.Add("Authorization", *httpAuth)
	}

	c.res.Reset()
	var start = time.Now()
	err = c.httpClient.Do(c.req, c.res)
	elapsed := time.Since(start).Seconds()
	if err != nil {
		return &gompet.ClientResult{Err: err, Time: elapsed}
	}
	status := c.res.StatusCode()
	resBody := c.res.Body()
	if status < 200 || status > 299 {
		log.Printf("%d fasthttp status %d body '%s'", c.config.ID, status, string(resBody))
	}
	elapsed = time.Since(start).Seconds() // final time will include body read time
	var res = fmt.Sprintf("%d %s", status, http.StatusText(status))
	if c.config.Verbose {
		log.Printf("%d fasthttp %s '%s' body '%s'", c.config.ID, cmd, res, string(resBody))
	}
	return &gompet.ClientResult{Res: res, Time: elapsed}
}

func (c *myClient) Term() {
	log.Println(c.config.ID, "fasthttp term")
}

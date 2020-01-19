// This file is part of Gompet - Copyright 2019-2020 Jari Karjala - www.jpkware.com
// SPDX-License-Identifier: GPLv3-only

// HTTP Client using Go standard library
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jkarjala/gompet"
)

// Long options for the testers, short ones used by the main library
var httpAuth = flag.String("auth", "", "HTTP Authorization header")
var httpContentType = flag.String("content-type", "application/json", "HTTP body content type")
var httpTimeout = flag.Int("timeout", 10, "HTTP Client timeout in seconds")

func main() {
	log.Println("HTTP tester started")
	gompet.Run(clientFactory)
}

type myClient struct {
	config     gompet.ClientConfig
	httpClient *http.Client
}

func clientFactory(config gompet.ClientConfig) (gompet.Client, error) {
	log.Println(config.ID, "http init")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConnsPerHost: 2, // each client has its own HTTP client
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}
	tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	httpClient := &http.Client{Transport: tr, Timeout: time.Duration(*httpTimeout) * time.Second}
	var client = myClient{config, httpClient}
	return &client, nil
}

func (c *myClient) RunCommand(in *gompet.ClientInput) *gompet.ClientResult {
	cmd := in.Cmd
	if c.config.Template != nil {
		cmd = c.config.Template.Expand(in.Args)
	}

	var req *http.Request
	var resp *http.Response
	var err error

	var ss = strings.SplitN(cmd, " ", 3)
	if len(ss) < 2 {
		return &gompet.ClientResult{Err: fmt.Errorf("Invalid command %s, HTTP verb and URL required", cmd)}
	}
	var primitive = strings.Trim(ss[0], "\r\t ")
	var url = strings.Trim(ss[1], "\r\t ")
	var body = ""
	if len(ss) == 3 {
		body = strings.Trim(ss[2], "\r\t")
	}

	req, err = http.NewRequest(primitive, url, strings.NewReader(body))
	if err != nil {
		return &gompet.ClientResult{Err: err}
	}

	if body != "" {
		req.Header.Add("Content-Type", *httpContentType)
	}
	if *httpAuth != "" {
		req.Header.Add("Authorization", *httpAuth)
	}

	var start = time.Now()
	resp, err = c.httpClient.Do(req)
	elapsed := time.Since(start).Seconds()
	if err != nil {
		return &gompet.ClientResult{Err: err, Time: elapsed}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		s := strings.ReplaceAll(string(body), "\n", " ")
		log.Printf("%d http response status %d body '%s'", c.config.ID, resp.StatusCode, s)
	} else {
		// body, _ := ioutil.ReadAll(resp.Body)
		// log.Printf("%s: %s\n", cmd, body)
		io.Copy(ioutil.Discard, resp.Body)
	}
	elapsed = time.Since(start).Seconds() // final time will include body read time
	var res = resp.Status
	if c.config.Verbose {
		log.Printf("%d http %s: %s", c.config.ID, cmd, res)
	}
	return &gompet.ClientResult{Res: res, Time: elapsed}
}

func (c *myClient) Term() {
	log.Println(c.config.ID, "http term")
}

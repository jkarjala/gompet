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

	"github.com/jkarjala/gomb"
)

// Long options for the testers, short ones used by the main library
var httpAuth = flag.String("auth", "", "HTTP Authorization header")
var httpContentType = flag.String("content-type", "application/json", "HTTP body content type")
var httpTimeout = flag.Int("timeout", 10, "HTTP Client timeout")

func main() {
	log.Println("HTTP tester started")
	gomb.Run(clientFactory)
}

type myClient struct {
	id         int
	template   *gomb.VarTemplate
	httpClient *http.Client
}

func clientFactory(id int, template string) (gomb.Client, error) {
	log.Println(id, "http init", template)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConnsPerHost: *gomb.NumClients,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}
	tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	httpClient := &http.Client{Transport: tr, Timeout: time.Duration(*httpTimeout) * time.Second}
	var client = myClient{id, gomb.Parse(template), httpClient}
	return &client, nil
}

func (c *myClient) RunCommand(in *gomb.RunInput) *gomb.RunResult {
	cmd := in.Cmd
	if c.template != nil {
		cmd = c.template.Expand(in.Args)
	}

	var req *http.Request
	var resp *http.Response
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

	req, err = http.NewRequest(primitive, url, strings.NewReader(body))
	if err != nil {
		return &gomb.RunResult{Err: err}
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
		return &gomb.RunResult{Err: err, Time: elapsed}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("%d HTTP status %d body '%s'", c.id, resp.StatusCode, string(body))
	} else {
		// body, _ := ioutil.ReadAll(resp.Body)
		// log.Printf("%s: %s\n", cmd, body)
		io.Copy(ioutil.Discard, resp.Body)
	}
	var res = resp.Status
	if *gomb.Verbose {
		log.Printf("%d http %s: %s", c.id, cmd, res)
	}
	return &gomb.RunResult{Res: res, Time: elapsed}
}

func (c *myClient) Term() {
	log.Println(c.id, "http term")
}

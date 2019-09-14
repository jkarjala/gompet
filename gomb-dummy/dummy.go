package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/jkarjala/gomb"
)

// Long options for the testers, short ones used by the main library
var delay = flag.Int("delay", 0, "Sleep in nanoseconds")

func main() {
	log.Println("Dummy tester started")
	gomb.Run(clientFactory)
}

type myClient struct {
	id       int
	template *gomb.VarTemplate
}

func clientFactory(id int, template string) (gomb.Client, error) {
	log.Println(id, "Dummy init", template)

	var client = myClient{id, gomb.Parse(template)}
	return &client, nil
}

func (c *myClient) RunCommand(in *gomb.RunInput) *gomb.RunResult {
	cmd := in.Cmd
	if c.template != nil {
		cmd = c.template.Expand(in.Args)
	}

	_ = cmd
	if *delay > 0 {
		time.Sleep(time.Duration(*delay))
	}

	if *gomb.Verbose {
		log.Println(cmd)
	}
	var res = fmt.Sprintf("%d OK", len(cmd))
	var elapsed = 0.0
	return &gomb.RunResult{Res: res, Time: elapsed}
}

func (c *myClient) Term() {
	log.Println(c.id, "Dummy term")
}

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jkarjala/gompet"
)

// Long options for the testers, short ones used by the main library
var nsDelay = flag.Int("ns-delay", 0, "Sleep in nanoseconds (real resolution depends on OS)")
var nsMax = flag.Float64("ns-max", 0, "Nanosecond max for random result times")
var nsSlow = flag.Float64("ns-add", 0, "Nanoseconds to add to result time delta after each cmd")

func main() {
	log.Println("Dummy tester started")
	gompet.Run(clientFactory)
}

type myClient struct {
	id       int
	template *gompet.VarTemplate
	delta    float64
}

func clientFactory(id int, template string) (gompet.Client, error) {
	log.Println(id, "Dummy init", template)

	var client = myClient{id, gompet.Parse(template), 0}
	return &client, nil
}

func (c *myClient) RunCommand(in *gompet.RunInput) *gompet.RunResult {
	cmd := in.Cmd
	if c.template != nil {
		cmd = c.template.Expand(in.Args)
	}

	_ = cmd
	if *nsDelay > 0 {
		time.Sleep(time.Duration(*nsDelay))
	}

	if *gompet.Verbose {
		log.Println(cmd)
	}
	var res = fmt.Sprintf("%d OK", len(cmd))
	var elapsed = *nsMax/1E9*rand.Float64() + c.delta
	c.delta += *nsSlow / 1E9
	return &gompet.RunResult{Res: res, Time: elapsed}
}

func (c *myClient) Term() {
	log.Println(c.id, "Dummy term")
}

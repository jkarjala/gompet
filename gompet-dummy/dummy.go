// This file is part of Gompet - Copyright 2019-2020 Jari Karjala - www.jpkware.com
// SPDX-License-Identifier: GPLv3-only

// Dummy test client for local testing
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
	config gompet.ClientConfig
	delta  float64
}

func clientFactory(config gompet.ClientConfig) (gompet.Client, error) {
	log.Println(config.ID, "Dummy init")

	var client = myClient{config, 0}
	return &client, nil
}

func (c *myClient) RunCommand(in *gompet.ClientInput) *gompet.ClientResult {
	cmd := in.Cmd
	if c.config.Template != nil {
		cmd = c.config.Template.Expand(in.Args)
	}

	_ = cmd
	if *nsDelay > 0 {
		time.Sleep(time.Duration(*nsDelay))
	}

	if c.config.Verbose {
		log.Println(cmd)
	}
	var res = fmt.Sprintf("%d OK", len(cmd))
	var elapsed = *nsMax/1E9*rand.Float64() + c.delta
	c.delta += *nsSlow / 1E9
	return &gompet.ClientResult{Res: res, Time: elapsed}
}

func (c *myClient) Term() {
	log.Println(c.config.ID, "Dummy term")
}

// Copyright 2019 Jari Karjala - www.jpkware.com

// Package gomb (Go Multi-purpose Benchmark) provides a multi-core benchmarking
// skeleton for use in different benchmarking clients. See sub-folders for some
// example command line clients.
package gomb

import (
	"bufio"
	"encoding/csv"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

var filename = flag.String("f", "-", "Input file name, stdin if not given or '-'")
var cmdTemplate = flag.String("t", "", "Command template, $1-$9 refers to input file columns (tab-separated)")
var numClients = flag.Int("c", 1, "Number of parallel clients executing commands")
var verbose = flag.Bool("v", false, "Verbose logging")

// RunInput is given to client once per input file line
type RunInput struct {
	Cmd  string   // command from file, nil if template
	Args []string // empty unless template is used
}

// RunResult is returned from client after processing one input line
type RunResult struct {
	Res  string  // result, count of each separate value is reported
	Time float64 // execution time in seconds, percentiles are reported
	Err  error   // error result or nil, count of each error is reported (if any)
}

// Client is the interface the client must implement
type Client interface {
	RunCommand(in RunInput) RunResult
	Term()
}

// ClientFactory generates a client instance
type ClientFactory func(id int, template string) (Client, error)

func init() {
	// Nothing to do now
}

// OpenInput opens the input file for reading
func OpenInput() *bufio.Reader {
	var err error
	var file = os.Stdin
	if *filename != "-" {
		file, err = os.Open(*filename)
		if err != nil {
			panic(err)
		}
	}
	return bufio.NewReader(file)
}

var clients []Client
var inputChan = make(chan RunInput)
var outputChan = make(chan RunResult)
var waitGroup sync.WaitGroup
var done = make(chan bool)

// Run function executes the benchmark and reports results
func Run(clientFactory ClientFactory) {
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
		log.SetFlags(0)
	}
	reader := OpenInput()

	err := LaunchClients(clientFactory)
	if err != nil {
		panic(err)
	}
	var results = NewResults()
	go CollectResults(results)

	if *cmdTemplate != "" {
		tsvReader := csv.NewReader(reader)
		FeedArgs(tsvReader)
	} else {
		FeedCmds(reader)
	}

	close(inputChan)
	log.Println("Waiting clients to finish")
	waitGroup.Wait()
	close(outputChan)

	log.Println("Waiting done from collect")
	<-done

	results.Report()
}

// LaunchClients creates clients and starts the go routines for data processing
func LaunchClients(clientFactory ClientFactory) error {
	clients = make([]Client, *numClients)
	var err error
	for i := 0; i < *numClients; i++ {
		clients[i], err = clientFactory(i, *cmdTemplate)
		if err != nil {
			return err
		}
		go ClientRoutine(i, clients[i])
	}
	return nil
}

// ClientRoutine is the processing function for data processing
func ClientRoutine(id int, client Client) {
	log.Printf("client %d started\n", id)
	waitGroup.Add(1)
	defer log.Printf("client %d exited\n", id)
	defer waitGroup.Done()

	for input := range inputChan {
		res := client.RunCommand(input)
		outputChan <- res
	}
	client.Term()
}

// CollectResults listens for processing results and updates the results
func CollectResults(results *Results) {
	log.Println("Waiting results")
	for res := range outputChan {
		results.Update(res)
	}
	log.Println("Results collected")
	close(done)
}

// FeedCmds feeds the clients with command lines
func FeedCmds(reader *bufio.Reader) error {
	log.Println("Feeding commands")
	for {
		cmd, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		cmd = strings.Trim(cmd, "\n")
		inputChan <- RunInput{Cmd: cmd}
	}
	return nil
}

// FeedArgs feeds the clients with arguments to patch to template
func FeedArgs(reader *csv.Reader) error {
	log.Println("Feeding args")
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Println("sending", row)
		inputChan <- RunInput{Args: row}
	}
	return nil
}

// Copyright 2019 Jari Karjala - www.jpkware.com

// Package gompet (Go Multi-purpose Performance Evaluation Tool) provides a multi-core
// benchmarking skeleton for use in different benchmarking clients.
// See sub-folders for some example command line clients.
package gompet

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http" // for profiler
	_ "net/http/pprof"
	"os"
	"strings"
	"sync"
)

var filename = flag.String("f", "", "Input file name, stdin if '-'")
var cmdTemplate = flag.String("t", "", "Command template, $1-$9 refers to tab-separated columns in input")
var progress = flag.Bool("P", false, "Report progress once a second")
var profile = flag.Bool("pprof", false, "Enable pprof web server")

var repeat = flag.Int("r", 1, "Repeat the input N times, does not work with stdin")

//var qps = flag.Int("R", 0, "Rate limit each client to N queries/sec")
//var snapshot = flag.Int("D", 0, "Display and discard percentiles every N minutes to save memory")

// NumClients exported for command implementations
var NumClients = flag.Int("c", 1, "Number of parallel clients executing commands")

// Verbose exported for command implementations
var Verbose = flag.Bool("v", false, "Verbose logging")

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
	RunCommand(in *RunInput) *RunResult
	Term()
}

// ClientFactory generates a client instance
type ClientFactory func(id int, template string) (Client, error)

func init() {
	// Nothing to do now
}

// OpenInput opens the input file for reading
func OpenInput() (io.Reader, *os.File) {
	var err error
	var file = os.Stdin
	if flag.NArg() > 0 {
		args := strings.Join(flag.Args(), "\n") + "\n"
		reader := strings.NewReader(args)
		return reader, nil
	}
	if *filename == "-" {
		return io.Reader(file), nil
	}
	if *filename != "" {
		file, err = os.Open(*filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return io.Reader(file), file
	}
	return nil, nil
}

var clients []Client
var inputChan = make(chan *RunInput)
var outputChan = make(chan *RunResult)
var waitGroup sync.WaitGroup
var done = make(chan bool)

// Run function executes the benchmark and reports results
func Run(clientFactory ClientFactory) {
	flag.Parse()
	if *profile {
		go func() {
			fmt.Println("Profiler active in http://localhost:4221/debug/pprof")
			http.ListenAndServe("localhost:4221", nil)
		}()
	}

	if !*Verbose {
		// it seems log still does sprintfs, check for *Verbose in loops
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	if *filename == "-" && *repeat > 1 {
		fmt.Println("Cannot use -r with stdin")
		os.Exit(1)
	}
	err := LaunchClients(clientFactory)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	var results = NewResults()
	go CollectResults(results)

	for loop := 0; loop < *repeat; loop++ {
		reader, file := OpenInput()
		if reader == nil {
			fmt.Println("Either 'command line' or -f filename must be given")
			os.Exit(1)
		}
		feedInput(reader, file)
	}
	close(inputChan)

	log.Println("Waiting clients to finish")
	waitGroup.Wait()
	close(outputChan)

	log.Println("Waiting done from collect")
	<-done

	results.Report(*progress)

	if *profile {
		fmt.Println("Run ready, ctrl-c to exit")
		select {} // wait forever
	}
}

func feedInput(reader io.Reader, file *os.File) {
	if file != nil {
		defer file.Close()
	}

	if *cmdTemplate != "" {
		tsvReader := csv.NewReader(reader)
		tsvReader.Comma = '\t'
		FeedArgs(tsvReader)
	} else {
		FeedCmds(reader)
	}
}

// LaunchClients creates clients and starts the go routines for data processing
func LaunchClients(clientFactory ClientFactory) error {
	clients = make([]Client, *NumClients)
	var err error
	for i := 0; i < *NumClients; i++ {
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
		results.Update(res, *progress)
	}
	log.Println("Results collected")
	close(done)
}

// FeedCmds feeds the clients with command lines
func FeedCmds(reader io.Reader) error {
	log.Println("Feeding commands")
	bufreader := bufio.NewReader(reader)
	for {
		cmd, err := bufreader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		cmd = strings.Trim(cmd, "\n")
		inputChan <- &RunInput{Cmd: cmd}
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
		if *Verbose {
			log.Println("sending", row)
		}
		inputChan <- &RunInput{Args: row}
	}
	return nil
}

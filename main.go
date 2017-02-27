package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

const version = "0.1.0"

var usestdin bool
var usejson bool
var usefile string
var verbose bool
var ehuri string
var usr string
var pass string

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func stderr(msg string) {
	fmt.Fprintf(os.Stderr, msg)
}

func init() {
	showVersion := flag.Bool(
		"V",
		false,
		"Show software version and exit",
	)
	flag.BoolVar(
		&usestdin,
		"c",
		false,
		"Read input data for API call from stdin",
	)
	flag.BoolVar(
		&usejson,
		"j",
		false,
		"Input and output are in JSON format",
	)
	flag.BoolVar(
		&verbose,
		"v",
		false,
		"Show full headers sent and received during API call",
	)
	flag.StringVar(
		&usefile,
		"f",
		"",
		"Read input data for API call from FILENAME",
	)
	flag.StringVar(
		&ehuri,
		"ehuri",
		os.Getenv("EHURI"),
		"Override ehuri env var",
	)
	ehauth := flag.String(
		"ehauth",
		os.Getenv("EHAUTH"),
		"Override ehauth env var",
	)
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	quit := false
	if ehuri == "" {
		stderr(
			"Please provide the <API endpoint URI> as an argument or EHURI " +
				"environment variable\n",
		)
		quit = true
	}
	if *ehauth == "" {
		stderr(
			"Please provide <user uuid>:<secret API key> as an argument or EHAUTH " +
				"environment variable\n",
		)
		quit = true
	} else {
		userpass := strings.Split(*ehauth, ":")
		if len(userpass) == 2 {
			usr, pass = userpass[0], userpass[1]
		} else {
			quit = true
			stderr("Malformed ehauth, please review it\n")
		}
	}
	if quit {
		os.Exit(1)
	}

	for ehuri[len(ehuri)-1:] == "/" {
		ehuri = ehuri[:len(ehuri)-1]
	}
}

func main() {
	var err error
	var stream io.Reader
	stream = nil
	uri := strings.Join(append([]string{ehuri}, flag.Args()...), "/")
	method := "GET"
	contentType := "text/plain"
	client := http.Client{}

	if usefile != "" {
		dat, err := ioutil.ReadFile(usefile)
		check(err)
		stream = strings.NewReader(string(dat))
		method = "POST"
		contentType = "application/octet-stream"
	} else if usestdin {
		var lines []string
		strm := bufio.NewReader(os.Stdin)
		x, err := strm.ReadString('\n')
		for err != io.EOF {
			check(err)
			lines = append(lines, x)
			x, err = strm.ReadString('\n')
		}
		stream = strings.NewReader(strings.Join(lines, ""))
		method = "POST"
		contentType = "application/octet-stream"
	}

	req, err := http.NewRequest(method, uri, stream)
	check(err)
	req.SetBasicAuth(usr, pass)
	if usejson {
		req.Header.Set("Accept", "application/json")
	}
	req.Header.Set("Content-Type", contentType)
	if verbose {
		dump, err := httputil.DumpRequestOut(req, true)
		check(err)
		fmt.Printf("Request Dump:\n%s\n\n", dump)
	}

	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()
	if verbose {
		dump, err := httputil.DumpResponse(resp, false)
		check(err)
		fmt.Printf("Response Dump:\n%s\n\n", dump)
	}
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(resp.Body)
	check(err)
	if n > 0 {
		fmt.Println(buf)
	}
}

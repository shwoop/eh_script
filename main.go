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

	if ehuri == "" {
		stderr(
			"Please provide the <API endpoint URI> as an argument or EHURI " +
				"environment variable\n",
		)
		defer os.Exit(1)
	}
	if *ehauth == "" {
		stderr(
			"Please provide <user uuid>:<secret API key> as an argument or EHAUTH " +
				"environment variable\n",
		)
		defer os.Exit(1)
	} else {
		userpass := strings.Split(*ehauth, ":")
		if len(userpass) == 2 {
			usr, pass = userpass[0], userpass[1]
		} else {
			stderr("Malformed ehauth, please review it\n")
			defer os.Exit(1)
		}
	}

	for ehuri[len(ehuri)-1:] == "/" {
		ehuri = ehuri[:len(ehuri)-1]
	}
}

func query_api() error {
	var (
		err error
		stream io.Reader
	)
	uri := strings.Join(append([]string{ehuri}, flag.Args()...), "/")
	method := "GET"
	contentType := "text/plain"
	client := http.Client{}

	if usefile != "" {
		dat, err := ioutil.ReadFile(usefile)
		if err != nil {
			return err
		}
		stream = strings.NewReader(string(dat))
		method = "POST"
		contentType = "application/octet-stream"
	} else if usestdin {
		var lines []string
		strm := bufio.NewReader(os.Stdin)
		x, err := strm.ReadString('\n')
		for err != io.EOF {
			if err != nil {
				return err
			}
			lines = append(lines, x)
			x, err = strm.ReadString('\n')
		}
		stream = strings.NewReader(strings.Join(lines, ""))
		method = "POST"
		contentType = "application/octet-stream"
	}

	req, err := http.NewRequest(method, uri, stream)
	if err != nil {
		return err
	}
	req.SetBasicAuth(usr, pass)
	if usejson {
		req.Header.Set("Accept", "application/json")
	}
	req.Header.Set("Content-Type", contentType)
	if verbose {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return err
		}
		fmt.Printf("Request Dump:\n%s\n\n", dump)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if verbose {
		dump, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return err
		}
		fmt.Printf("Response Dump:\n%s\n\n", dump)
	}
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	if n > 0 {
		fmt.Println(buf)
	}
	return nil
}

func main() {
	if err := query_api(); err != nil {
		stderr(err.Error())
		os.Exit(1)
	}
}

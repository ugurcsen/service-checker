package main

import (
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/ugurcsen/service-checker/opensearch"
	"github.com/ugurcsen/service-checker/types"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var serverCheckConfig types.ServerCheckConfig
var results map[string]types.ResultStruct
var wg sync.WaitGroup
var mu sync.Mutex
var verbose *bool
var openSearchActive bool

func main() {
	configFile := flag.String("c", "", "Config file")
	outputFile := flag.String("o", "", "Output file")
	timeInterval := flag.Int64("i", 0, "Time interval")
	verbose = flag.Bool("v", false, "Verbose mode")
	flag.Parse()

	if *configFile == "" {
		fmt.Println("Config file is required")
		os.Exit(1)
	}

	err := decodeConfig(*configFile)
	if err != nil {
		fmt.Println("Error reading config:", err)
		os.Exit(1)
	}

	if len(serverCheckConfig.Hosts) == 0 {
		fmt.Println("No hosts found")
		os.Exit(1)
	}

	if serverCheckConfig.OpenSearch != nil &&
		len(serverCheckConfig.OpenSearch.Hosts) > 0 &&
		serverCheckConfig.OpenSearch.Index != "" {
		openSearchActive = true
	} else {
		openSearchActive = false
	}

	for {
		results = make(map[string]types.ResultStruct, len(serverCheckConfig.Hosts))
		wg.Add(len(serverCheckConfig.Hosts))

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		http.DefaultTransport.(*http.Transport).DisableKeepAlives = true

		for _, prefixedIPAddress := range serverCheckConfig.Hosts {
			go checkServer(prefixedIPAddress)
		}

		wg.Wait()
		saveResults(*outputFile)
		if *timeInterval == 0 {
			break
		} else {
			time.Sleep(time.Second * time.Duration(*timeInterval))
		}
	}
}

func checkServer(prefixedIPAddress string) {
	defer wg.Done()
	ipAddress := strings.SplitN(prefixedIPAddress, "-", 2)[1]
	httpClient := &http.Client{
		Timeout:   time.Second * 30,
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
	}

	var port int
	var protocol, link string
	var ssrCheck bool
	var ssrThreshold int

	for k, v := range serverCheckConfig.Namespaces {
		if strings.HasPrefix(prefixedIPAddress, k) {
			port = v.Port
			protocol = v.Protocol
			link = v.Link
			ssrCheck = v.SSRCheck
			if v.SSRThreshold == 0 {
				ssrThreshold = 5 << 10 // 5KB
			}
			break
		}
	}

	if link == "" {
		fmt.Println("Namespace not found for " + prefixedIPAddress)
		os.Exit(1)
	}

	timeBefore := time.Now()
	url := fmt.Sprintf("%s://%s:%d%s", protocol, ipAddress, port, link)
	if *verbose == true {
		fmt.Println("Request url:", url)
	}
	resp, err := httpClient.Get(url)
	timeAfter := time.Now()
	latency := timeAfter.Sub(timeBefore)

	result := types.ResultStruct{
		Time: timeBefore,
		Host: prefixedIPAddress,
	}
	result.SetLatency(latency)

	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			result.StatusCode = "200"
			body, err := io.ReadAll(resp.Body)
			result.ContentLength = len(body)
			if ssrCheck {
				if err == nil {
					size := len(body)
					result.SSR = new(bool)
					if size > ssrThreshold {
						*result.SSR = true
					} else {
						*result.SSR = false
					}
				}
			}
		} else {
			result.StatusCode = fmt.Sprintf("%d", resp.StatusCode)
		}
	} else {
		result.StatusCode = err.Error()
	}

	mu.Lock()
	defer mu.Unlock()
	results[prefixedIPAddress] = result
}

func decodeConfig(filename string) error {
	var reader io.Reader
	if filename == "-" {
		reader = os.Stdin
		go func() {
			time.Sleep(5 * time.Second)
			os.Stdin.Close()
		}()
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		reader = file
	}

	decoder := yaml.NewDecoder(reader)

	for {
		err := decoder.Decode(&serverCheckConfig)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

func saveResults(outputFile string) {
	fmtRed := color.New(color.FgRed)
	fmtGreen := color.New(color.FgHiGreen)
	fmtBlue := color.New(color.FgBlue)

	var o *csv.Writer
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening output file:", err)
			os.Exit(1)
		}
		defer f.Close()
		o = csv.NewWriter(f)
		defer o.Flush()
	}
	for _, prefixedIPAddress := range serverCheckConfig.Hosts {
		r := results[prefixedIPAddress]

		str := prefixedIPAddress + " - Status: "

		if r.StatusCode == "200" {
			str += fmtGreen.Sprintf("%s", r.StatusCode)
		} else {
			str += fmtRed.Sprintf("%s", r.StatusCode)
		}
		ssrStr := "-"
		if r.SSR != nil {
			ssrStr = fmt.Sprintf("%v", *r.SSR)
			str += " - SSR: "
			if *r.SSR {
				str += fmtGreen.Sprintf("True")
			} else {
				str += fmtRed.Sprintf("False")
			}
		}

		str += fmtBlue.Sprintf(" - Latency: %f", r.GetLatency().Seconds())
		str += fmtBlue.Sprintf(" - ContentLength: %dKB", r.ContentLength>>10)
		fmt.Println(str)

		if o != nil {
			if openSearchActive {
				err := opensearch.SendToOpenSearch(serverCheckConfig, r)
				if *verbose && err != nil {
					log.Println(err)
				}
			}
			err := o.Write([]string{r.Time.String(), r.Host, r.StatusCode, fmt.Sprintf("%f", r.GetLatency().Seconds()), fmt.Sprintf("%d", r.ContentLength), ssrStr})
			if err != nil {
				fmt.Println("Error writing to output file:", err)
				os.Exit(1)
			}
		}
	}
}

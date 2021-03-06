package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"soilgaze/osint"
	"strings"
)

// OSINT interface to couple methods with respected .go files
type OSINT interface {
	Check(*[]osint.HostStruct)
}

// APIKeys : A struct that holds API key values
type APIKeys struct {
	Shodan     string `yaml:"api_shodan"`
	BinaryEdge string `yaml:"api_binaryedge"`
	Censys     string `yaml:"api_censys"`
	ZoomEyeU   string `yaml:"api_zoomeye_u"`
	ZoomEyeP   string `yaml:"api_zoomeye_p"`
	Onyphe     string `yaml:"api_onyphe"`
	Spyse      string `yaml:"api_spyse"`
}

func main() {
	var err error
	var allHosts []osint.HostStruct
	var apiKeys *APIKeys

	var hostFile string
	var configFile string
	var isEnvSet bool
	var osintList string
	var outFile string

	flag.StringVar(&hostFile, "host-file", "", "Location of the file that has a list of target hosts")
	flag.StringVar(&configFile, "config-file", "", "Location of the file that holds API key values in YAML format")
	flag.BoolVar(&isEnvSet, "config-env", false, "Should look at environment variables for API keys")
	flag.StringVar(&osintList, "osint-list", "", "OSINT resources to gather information from. Example: --osint-list=shodan,binaryedge")
	flag.StringVar(&outFile, "out-file", "", "File to write the results in JSON format. If not given, results will only be printed to console.")
	flag.Parse()

	if hostFile == "" {
		log.Fatal("A list of hosts should be provided!")
	}

	if isEnvSet {
		log.Println("Checking environment variables for API key values...")
		apiKeys, err = loadEnvironment()
	} else {
		if configFile == "" {
			log.Println("Config file location is not provided, will try to open default 'config.yaml' file...")
		}
		apiKeys, err = loadConfig(configFile)
	}

	if err != nil {
		log.Fatal(err)
	}

	hostsFileContents, err := readLines(hostFile)
	if err != nil {
		log.Fatal("An error occured while reading hosts file.")
	}

	prepareHostStruct(hostsFileContents, &allHosts)

	var shodan OSINT = osint.Shodan{APIKey: apiKeys.Shodan}
	var binaryedge OSINT = osint.Binaryedge{APIKey: apiKeys.BinaryEdge}
	var censys OSINT = osint.Censys{APIKey: apiKeys.Censys}
	var zoomeye OSINT = osint.Zoomeye{Username: apiKeys.ZoomEyeU, Password: apiKeys.ZoomEyeP}
	var onyphe OSINT = osint.Onyphe{APIKey: apiKeys.Onyphe}
	var spyse OSINT = osint.Spyse{APIKey: apiKeys.Spyse}

	if osintList == "" {
		shodan.Check(&allHosts)
		binaryedge.Check(&allHosts)
		censys.Check(&allHosts)
		zoomeye.Check(&allHosts)
		onyphe.Check(&allHosts)
		// spyse.check(&allHosts)
	} else {
		resources := strings.Split(osintList, ",")

		for _, name := range resources {
			if name == "shodan" {
				shodan.Check(&allHosts)
			} else if name == "binaryedge" {
				binaryedge.Check(&allHosts)
			} else if name == "censys" {
				censys.Check(&allHosts)
			} else if name == "zoomeye" {
				zoomeye.Check(&allHosts)
			} else if name == "onyphe" {
				onyphe.Check(&allHosts)
			} else if name == "spyse" {
				spyse.Check(&allHosts)
			}
		}
	}

	finalResult, err := json.Marshal(&allHosts)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(finalResult))

	if outFile != "" {
		if err := writeStringToFile(string(finalResult), outFile); err != nil {
			log.Fatalf("Could not write results to file: %s", err)
		}
	}
}

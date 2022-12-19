package main

import (
	"flag"
	"fmt"
	"os"

	"yuscan/Plugins"
	"yuscan/common"
)

func main() {
	var (
		model = flag.String("model", "subdomain", "model: subdomain,portscan")
		// subdomain
		flDomain      = flag.String("domain", "", "The domain to perform guessing against.")
		flWordlist    = flag.String("wordlist", "", "The wordlist to use for guessing.")
		flWorkerCount = flag.Int("c", 1000, "The amount of workers to use.")
		flServerAddr  = flag.String("server", "8.8.8.8:53", "The DNS server to use.")

		// portscan
		hostslistFile = flag.String("hostslist", "", "The hostslist to use for portscan.")
		timeout       = flag.Int64("timeout", 3, "The timeout to use for portscan.")
		ports         = flag.String("ports", "", "The ports to use for portscan.")
	)
	flag.IntVar(&common.Threads, "t", 600, "Thread nums")
	// flag.IntVar(&common.Proxy, "proxy", "http://127.0.0.1:10801", "http proxy")
	// flag.StringVar(&common.Socks5Proxy, "Socks5Proxy", "socks5://127.0.0.1:10800", "socks5 proxy")
	flag.Parse()

	if *model == "subdomain" {
		if *flDomain == "" || *flWordlist == "" {
			fmt.Println("-domain and -wordlist are required")
			os.Exit(1)
		}

		Plugins.Subdomain_guesser(flDomain, flWordlist, flWorkerCount, flServerAddr)

	}

	if *model == "portscan" {
		if *hostslistFile == "" || *ports == "" {
			fmt.Println("-domain and -ports is required")
			os.Exit(1)
		}

		Plugins.PortScan(hostslistFile, *ports, *timeout)

	}

}

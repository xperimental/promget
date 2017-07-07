package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	addr := pflag.StringP("address", "a", "http://localhost:8080/metrics", "Metrics endpoint to query.")
	interval := pflag.DurationP("interval", "i", 10*time.Second, "Scraping interval.")
	onlyList := pflag.BoolP("list", "l", false, "If true, only lists metric names.")
	pflag.Parse()

	if len(*addr) == 0 {
		log.Fatal("Need to provide an --address.")
	}

	if *interval < 0 {
		log.Fatal("Scraping --interval needs to be greater than zero.")
	}

	query := pflag.Arg(0)
	if !(*onlyList) && len(query) == 0 {
		log.Fatal("usage: promget [options] <query>")
	}

	scraper := newScraper(*addr, *interval)

	if *onlyList {
		samples, err := scraper.Scrape(time.Now())
		if err != nil {
			log.Fatalf("Error getting list: %s", err)
		}

		for _, s := range samples {
			name := s.Metric.String()
			fmt.Printf("%s\n", name)
		}
		return
	}

	scraper.Start()

	for sample := range scraper.Ch() {
		name := sample.Metric.String()
		if name == query {
			fmt.Printf("%d\n", int(sample.Value))
		}
	}
}

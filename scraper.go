package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

type scraper struct {
	client   *http.Client
	url      string
	interval time.Duration
	output   chan *model.Sample
	ticker   *time.Ticker
}

func newScraper(url string, interval time.Duration) *scraper {
	return &scraper{
		client:   &http.Client{},
		url:      url,
		interval: interval,
		output:   make(chan *model.Sample),
		ticker:   time.NewTicker(interval),
	}
}

func (s *scraper) Scrape(ts time.Time) (model.Samples, error) {
	res, err := s.client.Get(s.url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-ok status code: %d", res.StatusCode)
	}

	var (
		result     = make(model.Samples, 0, 200)
		decSamples = make(model.Vector, 0, 50)
	)
	sdec := expfmt.SampleDecoder{
		Dec: expfmt.NewDecoder(res.Body, expfmt.ResponseFormat(res.Header)),
		Opts: &expfmt.DecodeOptions{
			Timestamp: model.TimeFromUnixNano(ts.UnixNano()),
		},
	}

	for {
		if err = sdec.Decode(&decSamples); err != nil {
			break
		}
		result = append(result, decSamples...)
		decSamples = decSamples[:0]
	}

	return result, nil
}

func (s *scraper) Start() {
	go func() {
		for ts := range s.ticker.C {
			samples, err := s.Scrape(ts)
			if err != nil {
				log.Printf("Error during scrape: %s", err)
			}

			for _, sample := range samples {
				s.output <- sample
			}
		}
	}()
}

func (s *scraper) Ch() <-chan *model.Sample {
	return s.output
}

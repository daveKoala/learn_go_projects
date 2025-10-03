package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
)

var http string = "https://www.bbc.co.uk/news/articles/cvg885p923jo"

type logEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type logger struct {
	csvFile      *os.File
	jsonlFile    *os.File
	csvWriter    *csv.Writer
	jsonlEncoder *json.Encoder
}

func newLogger(csvPath, jsonlPath string) (*logger, error) {
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return nil, err
	}
	jsonlFile, err := os.Create(jsonlPath)
	if err != nil {
		csvFile.Close()
		return nil, err
	}

	return &logger{
		csvFile:      csvFile,
		jsonlFile:    jsonlFile,
		csvWriter:    csv.NewWriter(csvFile),
		jsonlEncoder: json.NewEncoder(jsonlFile),
	}, nil
}

func (l *logger) Close() error {
	l.csvWriter.Flush()
	if err := l.csvWriter.Error(); err != nil {
		return err
	}
	if err := l.csvFile.Close(); err != nil {
		return err
	}
	return l.jsonlFile.Close()
}

func (l *logger) Log(key, value string) error {
	entry := logEntry{Key: key, Value: value}
	if err := l.csvWriter.Write([]string{key, value}); err != nil {
		return err
	}
	l.csvWriter.Flush()
	if err := l.csvWriter.Error(); err != nil {
		return err
	}
	return l.jsonlEncoder.Encode(entry)
}

func main() {

	counter := 0

	logWriter, err := newLogger("scrape.csv", "scrape.jsonl")
	if err != nil {
		log.Fatalf("create loggers: %v", err)
	}
	defer func() {
		if err := logWriter.Close(); err != nil {
			log.Printf("close loggers: %v", err)
		}
	}()

	c := colly.NewCollector(
		colly.AllowedDomains("www.bbc.co.uk", "bbc.co.uk"),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"),
	)

	c.MaxDepth = 1

	// On every a element which has href attribute call callback
	// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	// 	link := e.Attr("href")
	// 	// Print link
	// 	fmt.Printf("Link found: %q -> %s\n", e.Text, link)
	// 	// Visit link found on page
	// 	// Only those links are visited which are in AllowedDomains
	// 	c.Visit(e.Request.AbsoluteURL(link))
	// })

	c.OnHTML("p", func(e *colly.HTMLElement) {
		if err := logWriter.Log("p", e.Text); err != nil {
			log.Printf("log paragraph: %v", err)
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		counter += 1
		fmt.Println("Visiting", r.URL.String())

		if counter == 100 {
			//
			fmt.Println("Reached 100 links, exiting...")
		}

	})

	// Start scraping on https://hackerspaces.org
	c.Visit(http)
}

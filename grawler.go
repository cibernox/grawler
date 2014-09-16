package main

import (
	"./urlinspector"
	"fmt"
	"net/http"
	"os"
)

type sitemapEntry struct {
	links  []string
	assets []string
}

type result struct {
	url string
	sitemapEntry
}

var fetchingFlag sitemapEntry = sitemapEntry{}

func InspectUrl(client urlinspector.Getter, url string, resultsChan chan *result) {
	fmt.Printf("...Fetching url %s\n", url)
	inspector := urlinspector.UrlInspector{Url: url, Client: client}
	links, err := inspector.Links()
	if err != nil {
		resultsChan <- nil
		return
	}
	assets, err := inspector.Assets()
	if err != nil {
		resultsChan <- nil
		return
	}
	entry := sitemapEntry{links, assets}
	resultsChan <- &result{url, entry}
}

func Crawl(client urlinspector.Getter, entryUrl string) map[string]*sitemapEntry {
	resultsChan := make(chan *result)
	sitemap := make(map[string]*sitemapEntry)

	pendingRequests := 1
	go InspectUrl(client, entryUrl, resultsChan)

	for result := range resultsChan {
		pendingRequests--
		if result != nil {
			sitemap[result.url] = &result.sitemapEntry
			for _, link := range result.links {
				if sitemap[link] == nil {
					sitemap[link] = &fetchingFlag
					pendingRequests++
					go InspectUrl(client, link, resultsChan)
				}
			}
		}
		if pendingRequests == 0 {
			close(resultsChan)
		}
	}

	return sitemap
}

func main() {
	sitemap := Crawl(&http.Client{}, os.Args[1])

	fmt.Println("{")
	for url, info := range sitemap {
		fmt.Printf("  %s: {\n    links: {%s} \n    assets: {%s}\n  },\n", url, info.links, info.assets)
	}
	fmt.Println("}")

}

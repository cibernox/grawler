package main

import (
	// "fmt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"os"
	"testing"
)

func buildResponse(filename string) http.Response {
	file, _ := os.Open(filename)
	return http.Response{Status: "200 OK", Body: file}
}

type FakeHttpClient struct {
	responses map[string]http.Response
}

func (client *FakeHttpClient) Get(url string) (*http.Response, error) {
	response := client.responses[url]
	return &response, nil
}

func TestGrawl(t *testing.T) {

	Convey("Crawl", t, func() {
		mockResponses := make(map[string]http.Response)
		mockResponses["https://ex.com"] = buildResponse("test_files/root.html")
		mockResponses["https://ex.com/section-1"] = buildResponse("test_files/section-1.html")
		mockResponses["https://ex.com/section-2"] = buildResponse("test_files/section-2.html")
		mockResponses["https://ex.com/section-1-1"] = buildResponse("test_files/section-1-1.html")
		mockResponses["https://www.ex.com/subdomain"] = buildResponse("test_files/subdomain.html")
		fakeClient := FakeHttpClient{mockResponses}

		Convey("crawls the entire structure of a page, collecting links and assets", func() {
			var urls []string
			var entry *sitemapEntry
			sitemap := Crawl(&fakeClient, "https://ex.com")

			entry = sitemap["https://ex.com"]
			urls = []string{"https://ex.com/section-1", "https://ex.com/section-2", "https://www.ex.com/subdomain"}
			So(entry.links, ShouldResemble, urls)

			urls = []string{"https://ex.com/app.js", "https://external.com/img1.jpg", "https://ex.com/styles.css"}
			for _, url := range urls {
				So(entry.assets, ShouldContain, url)
			}

			entry = sitemap["https://ex.com/section-1"]
			urls = []string{"https://ex.com/section-1-1"}
			So(entry.links, ShouldResemble, urls)

			urls = []string{"https://ex.com/styles.css"}
			for _, url := range urls {
				So(entry.assets, ShouldContain, url)
			}

			entry = sitemap["https://ex.com/section-2"]
			urls = []string{"https://ex.com/section-1-1"}
			So(entry.links, ShouldResemble, urls)

			urls = []string{"https://ex.com/styles.css"}
			for _, url := range urls {
				So(entry.assets, ShouldContain, url)
			}

			entry = sitemap["https://ex.com/section-1-1"]
			urls = []string{}
			So(entry.links, ShouldResemble, urls)
			So(entry.assets, ShouldResemble, []string{"https://ex.com/styles.css"})

			entry = sitemap["https://www.ex.com/subdomain"]
			urls = []string{}
			So(entry.links, ShouldResemble, urls)
			So(entry.assets, ShouldResemble, urls)

			// fmt.Printf("Result is %v", sitemap)
		})
	})
}

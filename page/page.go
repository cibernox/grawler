package page

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"regexp"
)

var contactRegex = regexp.MustCompile("^(mailto|tel):")
var binaryFileRegex = regexp.MustCompile("\x2E(pdf|jpg|jpeg|gif|svg|png|doc|docx)$")

func filterBlanks(ary []string) []string {
	result := make([]string, 0, len(ary))
	for _, e := range ary {
		if e != "" {
			result = append(result, e)
		}
	}
	return result
}

type Getter interface {
	Get(string) (*http.Response, error)
}

type Page struct {
	Url       string
	Client    Getter
	document  *goquery.Document
	parsedUrl *url.URL
}

func (page *Page) sanitizeUrl(rawUrl string, filterNonUrls bool) string {
	if contactRegex.MatchString(rawUrl) {
		return ""
	}

	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		fmt.Printf("Bad url %s. Skipping...\n", rawUrl)
		return ""
	}

	pageUrl := page.ParsedUrl()

	parsedUrl.Fragment = ""
	if parsedUrl.Scheme == "" {
		parsedUrl.Scheme = pageUrl.Scheme // this should be the page's url, but it is not parsed yet.
	}

	if parsedUrl.Host == "" {
		parsedUrl.Host = pageUrl.Host
	}

	if filterNonUrls {
		if parsedUrl.Host != pageUrl.Host {
			re := regexp.MustCompile(fmt.Sprintf(".*\x2E%s", pageUrl.Host))
			if !re.MatchString(parsedUrl.Host) {
				return ""
			}
		}
		if binaryFileRegex.MatchString(rawUrl) {
			return ""
		}
	}

	return parsedUrl.String()
}

func (page *Page) Document() *goquery.Document {
	if page.document == nil {
		response, err := page.Client.Get(page.Url)
		if err != nil {
			fmt.Printf("Error fetching url %s (%s). Skipping...\n", page.Url, err.Error())
			return nil
		}
		defer response.Body.Close()

		document, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			panic(fmt.Sprintf("Document cannot be parsed %s", err))
		}

		page.document = document
	}
	return page.document
}

func (page *Page) ParsedUrl() *url.URL {
	if page.parsedUrl == nil {
		parsedUrl, err := url.Parse(page.Url)
		if err != nil {
			fmt.Printf("Bad url %s. Skipping...\n", page.Url)
		}
		page.parsedUrl = parsedUrl
	}
	return page.parsedUrl
}

func (page *Page) Links() []string {
	doc := page.Document()
	if doc == nil {
		return nil
	}
	urls := doc.Find("a").Map(func(i int, s *goquery.Selection) string {
		href, _ := s.Attr("href")
		return page.sanitizeUrl(href, true)
	})

	return filterBlanks(urls)
}

func (page *Page) Assets() []string {
	doc := page.Document()
	if doc == nil {
		return nil
	}
	srcs := doc.Find("[src]").Map(func(i int, s *goquery.Selection) string {
		src, _ := s.Attr("src")
		return page.sanitizeUrl(src, false)
	})
	hrefs := doc.Find("link[href]").Map(func(i int, s *goquery.Selection) string {
		href, _ := s.Attr("href")
		return page.sanitizeUrl(href, false)
	})
	urls := append(srcs, hrefs...)

	return filterBlanks(urls)
}

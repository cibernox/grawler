package urlinspector

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/url"
	"regexp"
)

var contact = regexp.MustCompile("^(mailto|tel):")
var binaryFileRegex = regexp.MustCompile("\x2E(pdf|jpg|jpeg|gif|svg|png|doc|docx)$")

func filterBlanks(ary []string) []string {
	result := make([]string, 0, len(ary))
	for i := range ary {
		if e := ary[i]; e != "" {
			result = append(result, e)
		}
	}
	return result
}

type HtmlAnalyzer struct {
	url      *url.URL
	document *goquery.Document
}

func NewHtmlAnalyzer(rawUrl string, body *io.ReadCloser) *HtmlAnalyzer {
	url, err := url.Parse(rawUrl)
	if err != nil {
		panic(fmt.Sprintf("Bad URL %s", rawUrl))
	}

	document, err := goquery.NewDocumentFromReader(*body)
	if err != nil {
		panic(fmt.Sprintf("Document cannot be parsed %s", err))
	}

	return &HtmlAnalyzer{url, document}
}

func (analyzer *HtmlAnalyzer) Links() []string {
	urls := analyzer.document.Find("a").Map(func(i int, s *goquery.Selection) string {
		href, _ := s.Attr("href")
		return analyzer.sanitizeUrl(href, true)
	})

	return filterBlanks(urls)
}

func (analyzer *HtmlAnalyzer) Assets() []string {
	srcs := analyzer.document.Find("[src]").Map(func(i int, s *goquery.Selection) string {
		src, _ := s.Attr("src")
		return analyzer.sanitizeUrl(src, false)
	})
	hrefs := analyzer.document.Find("link[href]").Map(func(i int, s *goquery.Selection) string {
		href, _ := s.Attr("href")
		return analyzer.sanitizeUrl(href, false)
	})
	urls := append(srcs, hrefs...)

	return filterBlanks(urls)
}

func (analyzer *HtmlAnalyzer) sanitizeUrl(rawUrl string, filterNonUrls bool) string {
	if contact.MatchString(rawUrl) {
		return ""
	}

	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		fmt.Printf("Bad url %s. Skipping...\n", rawUrl)
		return ""
	}

	parsedUrl.Fragment = ""
	if parsedUrl.Scheme == "" {
		parsedUrl.Scheme = analyzer.url.Scheme
	}

	if parsedUrl.Host == "" {
		parsedUrl.Host = analyzer.url.Host
	}

	if filterNonUrls {
		if parsedUrl.Host != analyzer.url.Host {
			re := regexp.MustCompile(fmt.Sprintf(".*\x2E%s", analyzer.url.Host))
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

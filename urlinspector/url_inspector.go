package urlinspector

import (
	"fmt"
	"net/http"
)

type Getter interface {
	Get(string) (*http.Response, error)
}

type inspectorError struct {
	msg string
}

func (i *inspectorError) Error() string {
	return i.msg
}

type UrlInspector struct {
	Url      string
	Client   Getter
	analyzer *HtmlAnalyzer
}

func (inspector *UrlInspector) sanitizeUrl(url string) string {
	return url
}

func (inspector *UrlInspector) fetch() error {
	if inspector.analyzer != nil {
		return nil
	}

	response, err := inspector.Client.Get(inspector.Url)
	if err != nil {
		errorMsg := fmt.Sprintf("Error fetching url %s (%s). Skipping...", inspector.Url, err.Error())
		fmt.Println(errorMsg)
		return &inspectorError{errorMsg}
	}
	defer response.Body.Close()
	inspector.analyzer = NewHtmlAnalyzer(inspector.Url, &response.Body)
	return nil
}

func (inspector *UrlInspector) Links() ([]string, error) {
	err := inspector.fetch()
	if err != nil {
		return nil, err
	}
	return inspector.analyzer.Links(), nil
}

func (inspector *UrlInspector) Assets() ([]string, error) {
	err := inspector.fetch()
	if err != nil {
		return nil, err
	}
	return inspector.analyzer.Assets(), nil
}

package textextraction

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.sajari.com/docconv"
	"github.com/jaytaylor/html2text"
)

func URLToText(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil || resp == nil {
		return "", err
	}
	contentType := resp.Header.Get("Content-Type")
	contentType = strings.Split(contentType, ";")[0]
	b, _ := ioutil.ReadAll(resp.Body)
	return BytesToText(b, contentType)
}

func BytesToText(contents []byte, contentType string) (string, error) {
	if len(contents) <= 0 {
		return "", errors.New("empty content")
	}
	res, err := docconv.Convert(bytes.NewReader(contents), contentType, true)
	if err != nil || res == nil || len(res.Body) == 0 {
		//fallbacks
		switch contentType {
		case "text/html":
			return html2text.FromString(string(contents), html2text.Options{TextOnly: true})
		}
	}
	return res.Body, nil
}

func PathToText(path string) (string, error) {
	if len(path) <= 0 {
		return "", errors.New("empty path")
	}
	if strings.HasPrefix(path, "http") {
		return URLToText(path)
	}
	res, err := docconv.ConvertPath(path)
	if err != nil || res == nil {
		return "", err
	}
	return res.Body, nil
}

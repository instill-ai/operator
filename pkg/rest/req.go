package rest

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

type Request struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	RequestBody string            `json:"request_body"`
	Headers     map[string]string `json:"headers"`
}

type Response struct {
	StatusCode   int               `json:"status_code"`
	ResponseBody string            `json:"response_body"`
	Headers      map[string]string `json:"headers"`
}

func (r Request) sendReq() (Response, error) {
	req, _ := http.NewRequest(r.Method, r.URL, strings.NewReader(r.RequestBody))
	for h, v := range r.Headers {
		req.Header.Add(h, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	if resp == nil {
		return Response{}, errors.New("no response received")
	}
	b, _ := io.ReadAll(resp.Body)
	response := Response{
		StatusCode:   resp.StatusCode,
		ResponseBody: string(b),
	}
	response.Headers = make(map[string]string, len(resp.Header))
	for h, v := range resp.Header {
		response.Headers[h] = v[0]
	}
	return response, nil
}

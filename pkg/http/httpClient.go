package http

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type HttpClient struct {
	url string
}

type HttpResponse struct {
	StatusCode int
	Bytes      []byte
}

func CreateHttpClient(url string) *HttpClient {
	client := HttpClient{url: url}
	return &client
}

func (client *HttpClient) addHeaders(request *http.Request, headers map[string]string) {
	if headers != nil {
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}
}

func (client *HttpClient) RequestWithHeaders(requestType string, body string, headers map[string]string) HttpResponse {
	payload := strings.NewReader(body)

	request, _ := http.NewRequest(requestType, client.url, payload)

	client.addHeaders(request, headers)

	response, _ := http.DefaultClient.Do(request)

	httpResponse := HttpResponse{StatusCode: response.StatusCode}

	defer response.Body.Close()
	responseBytes, _ := ioutil.ReadAll(response.Body)
	httpResponse.Bytes = responseBytes

	return httpResponse
}

func (client *HttpClient) DeleteUrl(requestHeaders map[string]string, body string) HttpResponse {
	requestHeaders["Accepts"] = "application/json"

	return client.RequestWithHeaders("DELETE", body, requestHeaders)
}

func (client *HttpClient) GetUrl(requestHeaders map[string]string, body string) HttpResponse {
	requestHeaders["Accepts"] = "application/json"

	return client.RequestWithHeaders("GET", body, requestHeaders)
}

func (client *HttpClient) PostUrl(requestHeaders map[string]string, body string) HttpResponse {
	requestHeaders["Accepts"] = "application/json"

	return client.RequestWithHeaders("POST", body, requestHeaders)
}

func (client *HttpClient) PostUrlEncodedFormBody(body string) HttpResponse {
	requestHeaders := make(map[string]string)
	requestHeaders["content-type"] = "application/x-www-form-urlencoded"
	requestHeaders["cache-control"] = "no-cache"

	return client.RequestWithHeaders("POST", body, requestHeaders)
}

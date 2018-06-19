package main

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type HttpClient struct {
	url string
}

type HttpResponse struct {
	statusCode int
	bytes      []byte
}

func createHttpClient(url string) *HttpClient {
	client := HttpClient{url: url}
	return &client
}

func (client *HttpClient) post(body string) HttpResponse {
	return client.postWithHeaders(body, nil)
}

func (client *HttpClient) postWithHeaders(body string, headers map[string]string) HttpResponse {
	payload := strings.NewReader(body)

	request, _ := http.NewRequest("POST", client.url, payload)

	client.addHeaders(request, headers)

	response, _ := http.DefaultClient.Do(request)

	httpResponse := HttpResponse{statusCode: response.StatusCode}

	defer response.Body.Close()
	responseBytes, _ := ioutil.ReadAll(response.Body)
	httpResponse.bytes = responseBytes

	return httpResponse
}

func (client *HttpClient) postUrlEncodedFormBody(body string) HttpResponse {
	requestHeaders := make(map[string]string)
	requestHeaders["content-type"] = "application/x-www-form-urlencoded"
	requestHeaders["cache-control"] = "no-cache"

	return client.postWithHeaders(body, requestHeaders)
}

func (client *HttpClient) addHeaders(request *http.Request, headers map[string]string) {
	if headers != nil {
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}
}

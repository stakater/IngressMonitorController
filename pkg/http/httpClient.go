package http

import (
	"bytes"
	"io"
	"net/http"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("http-client")

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
	for key, value := range headers {
		request.Header.Add(key, value)
	}
}

func (client *HttpClient) RequestWithHeaders(requestType string, body []byte, headers map[string]string) HttpResponse {
	reader := bytes.NewReader(body)

	//   log.Info("NewRequest: METHOD: " + requestType + " URL: " + client.url + " PAYLOAD: " + string(body))

	request, err := http.NewRequest(requestType, client.url, reader)
	if err != nil {
		log.Error(err, "Failed to craft HTTP Request. METHOD: "+requestType+
			" URL: "+client.url+
			" PAYLOAD: "+string(body))
	}

	if headers != nil {
		client.addHeaders(request, headers)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(err, "")
	} else if response == nil {
		log.Error(nil, "got empty response")
	}

	httpResponse := HttpResponse{StatusCode: response.StatusCode}

	defer response.Body.Close()
	responseBytes, _ := io.ReadAll(response.Body)
	httpResponse.Bytes = responseBytes

	return httpResponse
}

func (client *HttpClient) DeleteUrl(requestHeaders map[string]string, body []byte) HttpResponse {
	return client.RequestWithHeaders("DELETE", body, requestHeaders)
}

func (client *HttpClient) GetUrl(requestHeaders map[string]string, body []byte) HttpResponse {

	return client.RequestWithHeaders("GET", body, requestHeaders)
}

func (client *HttpClient) PostUrl(requestHeaders map[string]string, body []byte) HttpResponse {
	return client.RequestWithHeaders("POST", body, requestHeaders)
}

func (client *HttpClient) PutUrl(requestHeaders map[string]string, body []byte) HttpResponse {
	return client.RequestWithHeaders("PUT", body, requestHeaders)
}

func (client *HttpClient) PostUrlEncodedFormBody(body string) HttpResponse {
	requestHeaders := make(map[string]string)
	requestHeaders["content-type"] = "application/x-www-form-urlencoded"
	requestHeaders["cache-control"] = "no-cache"

	return client.RequestWithHeaders("POST", []byte(body), requestHeaders)
}

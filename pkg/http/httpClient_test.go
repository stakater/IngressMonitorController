package http

import (
	"net/http"
	"testing"
)

func TestCreateHttpClient(t *testing.T) {
	url := "https://google.com"
	client := CreateHttpClient(url)

	if client.url != url {
		t.Error("Client URL should match the assigned url")
	}
}

func TestPostShouldReturn405Status(t *testing.T) {
	url := "https://google.com"
	client := CreateHttpClient(url)

	response := client.post("")

	if response.StatusCode != 405 {
		t.Error("Status code mismatch")
	}
}

func TestAddHeaders(t *testing.T) {
	url := "https://google.com"
	client := CreateHttpClient(url)

	request := &http.Request{}
	request.Header = http.Header{}

	headers := make(map[string]string)
	headers["Accepts"] = "application/json"

	client.addHeaders(request, headers)

	if value, ok := request.Header["Accepts"]; ok {
		if value[0] != "application/json" {
			t.Error("Accepts Header should have the value application/json")
		}
	} else {
		t.Error("Request should have the header Accepts")
	}
}

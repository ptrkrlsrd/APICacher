package acache

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestStorableResponse(t *testing.T) {
	status := "200 OK"
	testResponse := &http.Response{
		Status:     status,
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader("Hello world!")),
		Header:     http.Header{},
	}

	testResponse.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := NewStorableResponse(testResponse)
	if err != nil {
		t.Fail()
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

	if resp.Status != status {
		t.Fatalf("Expected status '200 ok' got %v", resp.Status)
	}

	if string(resp.Body) != "Hello world!" {
		t.Fatalf("Expected status 'Hello world!' got %v", resp.Body)
	}
}

func TestIfCanWriteStorableResponse(t *testing.T) {
	status := "200 OK"
	testResponse := &http.Response{
		Status:     status,
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader("Hello world!")),
		Header:     http.Header{},
	}

	testResponse.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := NewStorableResponse(testResponse)
	if err != nil {
		t.Fail()
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		t.Fail()
	}

	var readResp StorableResponse
	err = json.Unmarshal(respData, &readResp)
	if err != nil {
		t.Fail()
	}
}
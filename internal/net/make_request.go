package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func MakeRequest[Request any, Response any](url *url.URL, data Request) (*Response, error) {
	var response *Response
	encodedData, encodeJsonErr := json.Marshal(data)
	if encodeJsonErr != nil {
		return response, encodeJsonErr
	}

	//log.Println("encodedData", string(encodedData))

	httpResponse, httpErr := http.Post(url.String(), "application/json", bytes.NewBuffer(encodedData))
	if httpErr != nil {
		return response, httpErr
	}
	if httpResponse.StatusCode != http.StatusOK {
		return response, fmt.Errorf("MakeRequest: response status code is not OK: %s", httpResponse.Status)
	}

	defer httpResponse.Body.Close()

	respDump, err := httputil.DumpResponse(httpResponse, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("RESPONSE:\n%s", string(respDump))

	decodeJsonErr := json.NewDecoder(httpResponse.Body).Decode(&response)
	return response, decodeJsonErr
}

func MakeRequestWithEmptyResponse[Request any](url *url.URL, data Request) error {
	encodedData, encodeJsonErr := json.Marshal(data)
	if encodeJsonErr != nil {
		return encodeJsonErr
	}
	httpResponse, httpErr := http.Post(url.String(), "application/json", bytes.NewBuffer(encodedData))
	if httpErr != nil {
		return httpErr
	}
	if httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("MakeRequest: response status code is not OK: %s", httpResponse.Status)
	}
	return nil
}

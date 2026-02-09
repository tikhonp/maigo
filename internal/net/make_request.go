// Package net provides network request helpers.
package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func MakeRequest[Request any, Response any](url *url.URL, data Request) (*Response, error) {
	var response *Response
	encodedData, encodeJSONErr := json.Marshal(data)
	if encodeJSONErr != nil {
		return response, encodeJSONErr
	}

	// log.Println("encodedData", string(encodedData))

	httpResponse, httpErr := http.Post(url.String(), "application/json", bytes.NewBuffer(encodedData))
	if httpErr != nil {
		return response, httpErr
	}
	if httpResponse.StatusCode != http.StatusOK {
		return response, fmt.Errorf("MakeRequest: response status code is not OK: %s", httpResponse.Status)
	}

	defer httpResponse.Body.Close()

	// respDump, err := httputil.DumpResponse(httpResponse, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("RESPONSE:\n%s", string(respDump))

	decodeJSONErr := json.NewDecoder(httpResponse.Body).Decode(&response)
	return response, decodeJSONErr
}

func MakeRequestWithEmptyResponse[Request any](url *url.URL, data Request) error {
	encodedData, encodeJSONErr := json.Marshal(data)
	if encodeJSONErr != nil {
		return encodeJSONErr
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

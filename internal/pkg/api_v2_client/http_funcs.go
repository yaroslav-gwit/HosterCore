package ApiV2client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// A generic function to post to a URL on a Hoster node and (optionally) include a payload.
//
// - `auth` should be in the form of "user:password"
//
// - `url` should be the full URL to post to, aka "http://host:port/api/v2/host/settings/ssh-auth-key"
//
// - `inputPayload` (optional) should be a map of strings, aka `map[string]interface{}{"ssh_auth_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD..."}`
func PostFunc(url string, auth string, inputPayload ...map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), HTTP_CALL_TIMEOUT*time.Second)
	defer cancel()

	var req *http.Request
	var err error

	if len(inputPayload) > 0 {
		// if the payload is not empty, marshal it to json and create a new request with the payload
		jsonPayload, err := json.Marshal(inputPayload[0])
		if err != nil {
			return fmt.Errorf("error marshalling payload: %s", err.Error())
		}
		payload := strings.NewReader(string(jsonPayload))

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
		if err != nil {
			return fmt.Errorf("error creating new post request: %s", err.Error())
		}
	} else {
		// if the payload is empty, create a new request without a payload
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			return fmt.Errorf("error creating new post request: %s", err.Error())
		}
	}

	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+authEncoded)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error posting to: %s" + err.Error())
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error posting to: %s; err: %s; body: %s", url, err.Error(), strings.TrimSpace(string(body)))
	}
	if res.StatusCode > 299 || res.StatusCode < 200 {
		return fmt.Errorf("error posting to: %s; err: %d; body: %s", url, res.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// A generic function to send a "DELETE" request to a URL on a Hoster node and (optionally) include a payload.
//
// - `auth` should be in the form of "user:password"
//
// - `url` should be the full URL to post to, aka "http://host:port/api/v2/host/settings/ssh-auth-key"
//
// - `inputPayload` (optional) should be a map of strings, aka `map[string]interface{}{"ssh_auth_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD..."}`
func DeleteFunc(url string, auth string, inputPayload ...map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), HTTP_CALL_TIMEOUT*time.Second)
	defer cancel()

	var req *http.Request
	var err error

	if len(inputPayload) > 0 {
		// if the payload is not empty, marshal it to json and create a new request with the payload
		jsonPayload, err := json.Marshal(inputPayload[0])
		if err != nil {
			return fmt.Errorf("error marshalling payload: %s", err.Error())
		}
		payload := strings.NewReader(string(jsonPayload))

		req, err = http.NewRequestWithContext(ctx, http.MethodDelete, url, payload)
		if err != nil {
			return fmt.Errorf("error creating new delete request: %s", err.Error())
		}
		req.Header.Add("Content-Type", "application/json")
	} else {
		// if the payload is empty, create a new request without a payload
		req, err = http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
		if err != nil {
			return fmt.Errorf("error creating new delete request: %s", err.Error())
		}
	}

	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+authEncoded)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error posting to: %s" + err.Error())
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error deleting on: %s; err: %s; body: %s", url, err.Error(), strings.TrimSpace(string(body)))
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("error : %s; err: %d; body: %s", url, res.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// A generic function to send a "GET" request to a URL on a Hoster node.
//
// - `auth` should be in the form of "user:password"
//
// - `url` should be the full URL to GET from, aka "http://host:port/api/v2/host/settings/ssh-auth-key"
func GetFunc(url string, auth string) (r []byte, e error) {
	ctx, cancel := context.WithTimeout(context.Background(), HTTP_CALL_TIMEOUT*time.Second)
	defer cancel()

	var req *http.Request
	var err error

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		e = fmt.Errorf("error creating a new GET request: %s", err.Error())
		return
	}

	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+authEncoded)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		e = fmt.Errorf("error GETting from: %s" + err.Error())
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		e = fmt.Errorf("error GETting from: %s; err: %s; body: %s", url, err.Error(), strings.TrimSpace(string(body)))
		return
	}
	if res.StatusCode > 299 {
		e = fmt.Errorf("error : %s; err: %d; body: %s", url, res.StatusCode, strings.TrimSpace(string(body)))
		return
	}

	r = body
	return
}

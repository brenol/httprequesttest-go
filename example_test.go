package example_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sawadashota/httpclient-test"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func client(t *testing.T) *http.Client {
	t.Helper()

	body := example.ResponseBody{
		Text: "hello",
	}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	return NewTestClient(func(req *http.Request) *http.Response {
		if req.Header.Get("Authorization") != "Bearer ok_token" {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       nil,
				Header:     make(http.Header),
			}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(b)),
			Header:     make(http.Header),
		}
	})
}

func errorClient(t *testing.T) *http.Client {
	t.Helper()
	return NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
			Header:     make(http.Header),
		}
	})
}

func unexpectedBodyClient(t *testing.T) *http.Client {
	t.Helper()
	return NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("bad")),
			Header:     make(http.Header),
		}
	})
}

func TestApi_Do(t *testing.T) {
	cases := map[string]struct {
		token                string
		client               *http.Client
		expectHasError       bool
		expectedErrorMessage string
		expectedText         string
	}{
		"normal": {
			token:          "ok_token",
			client:         client(t),
			expectHasError: false,
			expectedText:   "hello",
		},
		"invalid token": {
			token:                "invalid_token",
			client:               client(t),
			expectHasError:       true,
			expectedErrorMessage: "bad response status code 401",
		},
		"internal server error response": {
			token:                "ok_token",
			client:               errorClient(t),
			expectHasError:       true,
			expectedErrorMessage: "bad response status code 500",
		},
		"unexpected body response": {
			token:                "ok_token",
			client:               unexpectedBodyClient(t),
			expectHasError:       true,
			expectedErrorMessage: "invalid character 'b' looking for beginning of value",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			e := example.New(c.token, example.OptionHTTPClient(c.client))
			resp, err := e.Do()

			if c.expectHasError {
				if err == nil {
					t.Errorf("expected error but no errors ouccured")
				}

				if err.Error() != c.expectedErrorMessage {
					t.Errorf("unexpected error message. expected %s, actual %s", c.expectedErrorMessage, err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf(err.Error())
				return // because when has error, response is nil
			}

			if resp.Text != c.expectedText {
				t.Errorf("unexpected response's text. expected %s, actual %s", c.expectedText, resp.Text)
			}
		})
	}
}

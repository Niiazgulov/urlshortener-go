package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testkeymap = map[string]string{}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	if method == "POST" {
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		defer resp.Body.Close()
		return resp.StatusCode, string(respBody)
	} else {
		respBody := resp.Header.Get("Location")
		require.NoError(t, err)
		defer resp.Body.Close()
		return resp.StatusCode, string(respBody)
	}
}

func TestRouter(t *testing.T) {
	r := NewRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()
	statusCode, body := testRequest(t, ts, "POST", "/")
	testkeymap[body] = ts.URL
	assert.Equal(t, http.StatusCreated, statusCode)
	statusCode, body = testRequest(t, ts, "GET", "/{id}")
	original := testkeymap[body]
	assert.Equal(t, http.StatusTemporaryRedirect, statusCode)
	assert.Equal(t, original, body)
}

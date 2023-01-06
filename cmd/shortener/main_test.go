package main

import (

	// "encoding/json"
	// "math/rand"
	// "bufio"
	// "strconv"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	// "io"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	// "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBestHandlerEver(t *testing.T) {
	var keymap = map[string]string{}

	type want struct {
		statusCode  int
		originalurl string
	}
	tests := []struct {
		name        string
		originalurl string
		short       string
		request     string
		want        want
	}{
		{
			name:        "test #1 POST",
			originalurl: "google.com",
			request:     http.MethodPost,
			want: want{
				statusCode: 201,
			},
		},
		{
			name:    "test #2 GET",
			request: http.MethodGet,
			//short: keymap[originalurl],
			want: want{
				originalurl: "google.com",
				statusCode:  307,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := "http://localhost:8080/"

			if tt.request == http.MethodPost {
				data := url.Values{}
				data.Set("url", tt.originalurl)
				request, err := http.NewRequest(tt.request, endpoint, bytes.NewBufferString(data.Encode()))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				w := httptest.NewRecorder()
				h := http.HandlerFunc(BestHandlerEver)
				h(w, request)
				result := w.Result()
				shorturlByte, err := ioutil.ReadAll(result.Body)
				//shorturl2 := strings.ReplaceAll(string(shorturlByte), "http://localhost:8080/", "")
				shorturl2 := string(shorturlByte)
				keymap[tt.originalurl] = shorturl2
				assert.Equal(t, tt.want.statusCode, result.StatusCode)
				//assert.Equal(t, keymap[tt.originalurl], "sdsdvsdv")
				require.NoError(t, err)
				err = result.Body.Close()
				require.NoError(t, err)
			}

			if tt.request == http.MethodGet {
				data := url.Values{}
				for key := range keymap {
					data.Set("url", keymap[key])
				}
				//assert.Equal(t, data.Get("url"), "asdasdasdasdasdasdasd")
				request, err := http.NewRequest(tt.request, data.Get("url"), nil)
				// request, err := http.NewRequest(tt.request, endpoint, bytes.NewBufferString(data.Encode()))
				// request := httptest.NewRequest(http.MethodPost, tt.request, nil)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				w := httptest.NewRecorder()
				h := http.HandlerFunc(BestHandlerEver)
				h(w, request)
				result := w.Result()
				longurl2 := result.Header.Get("Location")
				//longurl3 := longurl2.Header().Get("Location")
				//longurl := w.Header().Get("Location")
				assert.Equal(t, tt.want.statusCode, result.StatusCode)
				assert.Equal(t, tt.want.originalurl, longurl2)
				require.NoError(t, err)
				err = result.Body.Close()
				require.NoError(t, err)
			}
		})
	}
}

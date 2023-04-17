// Пакет handlers, описание в файле doc.go
package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const testLongURL = "test.ru"

func Example() {

	// Test PostHandler
	endpoint := "http://localhost:8080/"
	dataPOST := url.Values{}
	dataPOST.Set("url", testLongURL)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBufferString(dataPOST.Encode()))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(dataPOST.Encode())))

	// client := &http.Client{}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	shortID := string(body)

	// Test GetHandler
	req, err = http.NewRequest(http.MethodGet, shortID, nil)
	res, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer res.Body.Close()
	originalURL := res.Header.Get("Location")
	fmt.Println(originalURL)
}

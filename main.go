package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	url = flag.String("url", "", "URL of video to download")
)

const (
	limit = 1000
)

func main() {
	flag.Parse()

	resp, err := http.Head(*url)
	if err != nil {
		panic(err)
	}

	contentLength, err := strconv.Atoi(resp.Header["Content-Length"][0])
	length := contentLength / limit
	body := []byte{}
	diff := contentLength % limit

	for i := 0; i < limit; i++ {
		min := length * i
		max := length * (i + 1)

		if i == limit-1 {
			max += diff
		}

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, *url, nil)
		if err != nil {
			panic(err)
		}

		rangeHeader := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1)
		req.Header.Add("Range", rangeHeader)
		fmt.Printf("Sending request; URL: %s, Range: %s \n", *url, rangeHeader)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err)
		}

		body = append(body, data...)
	}

	ioutil.WriteFile("out", body, 0x777)
}

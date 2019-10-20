package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	url       = flag.String("url", "", "URL of video to download")
	chunkSize = flag.Int("chunkSize", 2, "Chunk size in MB")
)

func main() {
	flag.Parse()

	if *chunkSize < 1 {
		panic(fmt.Errorf("chunkSize must be at least 1"))
	}
	*chunkSize = *chunkSize * 1000000

	resp, err := http.Head(*url)
	if err != nil {
		panic(err)
	}

	contentLength, err := strconv.Atoi(resp.Header["Content-Length"][0])
	chunks := contentLength / *chunkSize
	body := []byte{}
	diff := contentLength % *chunkSize

	fmt.Println("Content-Length", contentLength)
	fmt.Println("URL:", *url)
	fmt.Println("Chunk Size:", *chunkSize)
	fmt.Println("Chunks:", chunks)
	fmt.Println("Diff:", diff)

	for i := 0; i < chunks; i++ {
		min := *chunkSize * i
		max := *chunkSize * (i + 1)

		if i == chunks-1 {
			max += diff
		}

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, *url, nil)
		if err != nil {
			panic(err)
		}

		rangeHeader := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1)
		req.Header.Add("Range", rangeHeader)
		fmt.Printf("Sending request; Range: %s \n", rangeHeader)
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

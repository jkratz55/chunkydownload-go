package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	downloadURL = flag.String("url", "", "URL of video to download")
	chunkSize   = flag.Int("chunkSize", 2, "Chunk size in MB")
	outFileName = flag.String("outFileName", "", "Name of saved download file on disk")
)

var (
	client = http.Client{}
	mutex  = sync.Mutex{}
	wg     sync.WaitGroup
)

func main() {

	flag.Parse()

	if *chunkSize < 1 {
		panic(fmt.Errorf("chunkSize must be at least 1"))
	}
	*chunkSize = *chunkSize * 1000000

	start := time.Now()
	resp, err := http.Head(*downloadURL)
	if err != nil {
		panic(err)
	}

	contentLength, err := strconv.Atoi(resp.Header["Content-Length"][0])
	chunks := contentLength / *chunkSize
	containers := []responseContainer{}
	diff := contentLength % *chunkSize

	fmt.Println("Content-Length", contentLength)
	fmt.Println("URL:", *downloadURL)
	fmt.Println("Chunk Size:", *chunkSize)
	fmt.Println("Chunks:", chunks)
	fmt.Println("Diff:", diff)

	for i := 0; i < chunks; i++ {
		min := *chunkSize * i
		max := *chunkSize * (i + 1)

		if i == chunks-1 {
			max += diff
		}

		req, err := http.NewRequest(http.MethodGet, *downloadURL, nil)
		if err != nil {
			panic(err)
		}
		rangeHeader := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1)
		req.Header.Add("Range", rangeHeader)

		wg.Add(1)
		go func(r *http.Request, idx int) {
			defer wg.Done()
			fmt.Printf("Downloading byte range %s \n", req.Header.Get("Range"))
			re, err := download(r, idx)
			if err != nil {
				panic(err)
			}
			mutex.Lock()
			containers = append(containers, re)
			mutex.Unlock()
		}(req, i)
	}

	wg.Wait()
	contents := buildFileContents(containers)

	var fname string
	if outFileName != nil && *outFileName != "" {
		fname = *outFileName
	} else {
		fname = filenameFromURL(*downloadURL)
	}

	err = ioutil.WriteFile(fname, contents, 0x777)
	if err != nil {
		panic(err)
	}

	elapsed := time.Since(start)
	fmt.Println("Downloaded took", elapsed.Seconds(), "seconds")
}

func buildFileContents(containers []responseContainer) []byte {
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].index < containers[j].index
	})
	data := []byte{}
	for _, d := range containers {
		data = append(data, d.data...)
	}
	return data
}

// filenameFromUrl extracts a filename from a given URL
func filenameFromURL(inputURL string) string {
	u, err := url.Parse(inputURL)
	if err != nil {
		panic(err)
	}
	u.RawQuery = ""
	return path.Base(u.Path)
}

// download retrieves a resource over HTTP
func download(req *http.Request, idx int) (responseContainer, error) {
	resp, err := client.Do(req)
	if err != nil {
		return responseContainer{}, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return responseContainer{}, err
	}

	return responseContainer{index: idx, data: data}, nil
}

// responseContainer is a structure to represent a partial response along with
// an index to track the order. This allows a file to be downloaded in chunks
// concurrently and then re-assembled in the correct order later by the index.
type responseContainer struct {
	index int
	data  []byte
}

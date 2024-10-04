package pcm

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

/*
asynchronous function that makes a HTTP request to the intel PCM
sensor server to retrieve sensor data.
*/
func makePCMRequest(url string, filename string, wg *sync.WaitGroup, readyChan chan struct{}) {
	text := ""

	defer wg.Done()
	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error creating request for %s: %s", url, err)
		return
	}

	// Add the custom header
	req.Header.Add("Accept", "application/json")

	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: 1 * time.Second}

	// Send the first request for metrics
	body, err := handleClientReq(url, req, client)
	if err != nil {
		readyChan <- struct{}{}
		return
	}
	text += string(body)
	readyChan <- struct{}{}

	// TODO: ensure channels are handled more safely
	// wait for ROI to finish
	<-readyChan
	// Send the second request
	body, err = handleClientReq(url, req, client)
	if err != nil {
		return
	}
	text += string(body)

	//write to string
	writeToFile(filename, text)

}

func handleClientReq(url string, req *http.Request, client *http.Client) (res string, err error) {
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending HTTP request.", err)
		return "", err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading HTTP response. ", err)
		return "", err
	}
	resp.Body.Close()
	return string(body), nil
}

func writeToFile(filename string, text string) {
	//write to string
	f, err := os.Create(filename)
	if err != nil {
		log.Errorf("Error creating file %s. %s", filename, err)
		return
	}
	defer f.Close()
	nbytes, err := f.WriteString(text)
	if err != nil {
		log.Error("Failed to write text to file! ", err)
		return
	}
	log.Infof("Wrote %d bytes to file %s.\n", nbytes, filename)

}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	var numClients, requestsPerClient int
	var proxyAddr string
	flag.IntVar(&numClients, "clients", 10, "number of clients")
	flag.IntVar(&requestsPerClient, "requests", 5, "number of requests per client")
	flag.StringVar(&proxyAddr, "proxy", "http://localhost:8080", "proxy address")
	flag.Parse()

	wg := new(sync.WaitGroup)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			for j := 0; j < requestsPerClient; j++ {
				start := time.Now()

				reqURL := fmt.Sprintf("%s/api/data?client=%d&req=%d", proxyAddr, clientID, j)

				req, err := http.NewRequest("GET", reqURL, nil)
				if err != nil {
					log.Printf("client %d: error creating request %d: %v", clientID, j, err)
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					log.Printf("client %d: request %d failed: %v", clientID, j, err)
					continue
				}
				defer resp.Body.Close()

				duration := time.Since(start)
				log.Printf("client %d: request %d completed in %v - status: %d",
					clientID, j, duration, resp.StatusCode)

				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	log.Println("all clients completed")
}

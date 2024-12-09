package pkg

import (
	"net/http"
	"time"
)

type DomainClient struct {
	domain   string
	throttle time.Duration
	retries  int
	backoff  time.Duration
	client   *http.Client
}

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

func NewDomainClient(domain string, throttle time.Duration, retries int, backoff time.Duration) *DomainClient {
	return &DomainClient{
		domain:   domain,
		throttle: throttle,
		retries:  retries,
		backoff:  backoff,
		client:   &http.Client{},
	}
}

func (dc *DomainClient) fetchPage(page string) int {
	var resp *http.Response
	var err error

	for i := 0; i <= dc.retries; i++ {
		req, _ := http.NewRequest("GET", page, nil)
		req.Header.Set("User-Agent", userAgent)

		resp, err = dc.client.Do(req)
		if err == nil {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				time.Sleep(dc.throttle)

				continue
			}

			return resp.StatusCode
		}

		time.Sleep(dc.backoff * time.Duration(1<<(i*2)))
	}

	return 0
}

type Result struct {
	URL  string
	Code int
}

func (dc *DomainClient) FetchPages(urls []string) []Result {
	results := make([]Result, 0, len(urls))

	urlChannel := make(chan string, 3)
	resultChannel := make(chan Result, 3)

	for w := 1; w <= 3; w++ {
		go func(urlChannel <-chan string, resultChannel chan<- Result) {
			for url := range urlChannel {
				resultChannel <- Result{URL: url, Code: dc.fetchPage(url)}
			}
		}(urlChannel, resultChannel)
	}

	go func() {
		for _, url := range urls {
			time.Sleep(dc.throttle)
			urlChannel <- url
		}
		close(urlChannel)
	}()

	for a := 1; a <= len(urls); a++ {
		results = append(results, <-resultChannel)
	}

	return results
}

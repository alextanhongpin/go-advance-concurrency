package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func retry(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if s, ok := err.(stop); ok {
			// Return the original error for later checking
			return s.error
		}

		if attempts--; attempts > 0 {
			// Add some randomness to prevent the Thundering Herd
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2
			time.Sleep(sleep)
			return retry(attempts, 2*sleep, fn)
		}
		return err
	}
	return nil
}

type stop struct {
	error
}

// Example usage
func DeleteThing(id string) error {
	// Build the request
	req, err := http.Request("DELETE", fmt.Sprintf("https://unreliable-api/%s", id), nil)
	if err != nil {
		return errors.New("unable to make request")
	}

	// Execute the request

	return retry(3, time.Second, func() error {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// This will result in a retry
			return err
		}

		defer resp.Body.Close()

		s := resp.StatusCode {
		case s >= 500:
			// Retry
			return fmt.Errorf("server error: %v", s)
		case s >= 400: 
			// Don't retry, it was the client's fault
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			// Happy
			return nil
		}
	})
}

func main() {

}

package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", LazyServer)
	http.ListenAndServe(":1111", nil)
}

func LazyServer(w http.ResponseWriter, r *http.Request) {
	headOrTails := rand.Intn(2)

	if headOrTails == 0 {
		time.Sleep(6 * time.Second)
		fmt.Fprintf(w, "Go! Slow server %v", headOrTails)
		fmt.Printf("Go! Slow server %v", headOrTails)
		return
	}

	fmt.Fprintf(w, "Go! Quick server %v", headOrTails)
	fmt.Printf("Go! Quick server %v", headOrTails)
	return
}

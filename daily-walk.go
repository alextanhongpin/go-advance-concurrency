// http://whipperstacker.com/2015/10/05/3-trivial-concurrency-exercises-for-the-confused-newbie-gopher/
package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// http://whipperstacker.com/2015/10/05/3-trivial-concurrency-exercises-for-the-confused-newbie-gopher/
func main() {

	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup

	alice := func() {
		log.Println("Alice started getting ready")
		duration := 60 + rand.Intn(30)
		time.Sleep(time.Second * time.Duration(duration))
		log.Printf("Alice spent %v seconds getting ready\n", duration)
		wg.Done()
	}

	bob := func() {
		log.Println("Bob started getting ready")
		duration := 60 + rand.Intn(30)
		time.Sleep(time.Duration(duration) * time.Second)
		log.Printf("Bob spent %v seconds getting ready\n", duration)
		wg.Done()
	}

	putOnShoes := func(name string) {
		log.Printf("%s started putting on shoes\n", name)
		duration := 35 + rand.Intn(10)
		time.Sleep(time.Duration(duration) * time.Second)
		log.Printf("%s spend %d seconds putting on shoes\n", name, duration)
		wg.Done()
	}

	startAlarm := func() {
		log.Println("Arming alarm")
		log.Println("Alarm is counting down...")
		go putOnShoes("Bob")
		go putOnShoes("Alice")
		time.Sleep(60 * time.Second)
		log.Println("Exiting and locking the door.")
		log.Println("Alarm is armed.")
		wg.Done()
	}

	wg.Add(2)

	log.Println("Let's go for a walk")

	go bob()
	go alice()

	wg.Wait()
	log.Println("Alarm is armed.")

	wg.Add(3)

	go startAlarm()

	wg.Wait()

}

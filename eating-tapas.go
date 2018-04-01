package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// http://whipperstacker.com/2015/10/05/3-trivial-concurrency-exercises-for-the-confused-newbie-gopher/
// Solution 2: Eating Tapas
func main() {
	var wg sync.WaitGroup

	rand.Seed(time.Now().UnixNano())

	people := []string{"Alice", "Bob", "Charlie", "Dave"}
	menus := []string{"chorizo", "chopitos", "pimientos de padron", "croquetas", "patatas bravas"}

	var dishes []string
	for i := 0; i < len(menus); i++ {
		numDish := 5 + rand.Intn(5)
		for j := 0; j < numDish; j++ {
			dishes = append(dishes, menus[i])
		}
	}

	shuffledDishes := shuffle(dishes)

	ch := make(chan interface{}, len(people))
	eat := func(person, menu string) {
		duration := 0 + rand.Intn(3)
		log.Printf("%s is enjoying some %s\n", person, menu)
		time.Sleep(time.Duration(duration) * time.Second)
		wg.Done()
		<-ch
	}

	wg.Add(len(shuffledDishes))

	log.Println("Bon appetit!")
	for i, dish := range shuffledDishes {
		ch <- true
		go func(person, dish string) {
			eat(person, dish)
		}(people[i%4], dish)
	}

	wg.Wait()

	log.Println("That was delicious!")
}

func shuffle(src []string) []string {
	out := make([]string, len(src))
	perm := rand.Perm(len(src))
	for i, v := range perm {
		out[v] = src[i]
	}
	return out
}

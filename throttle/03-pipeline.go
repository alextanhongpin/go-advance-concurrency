package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

type Pipes struct {
	Input  <-chan interface{}
	Output chan<- interface{}

	// Chuck errors to here:
	Err <-chan error

	// All Pipes will listen for a quit signal, if you receive this, gracefully quit the function
	// Pipes *should* listen to this when receiving *or* sending to Input, Output respectively.
	Quit <-chan struct{}
	// All Pipes will signal this waitGroup just before quiting the function.
	Done *sync.WaitGroup
}

// This will be refered as the first stage
func walkFiles(p Pipes) {
	defer p.Done.Done()

	for {
		select {
		case root_ := <-p.Input:
			if root_ == nil {
				// Channel is closed:
				return
			}
			root := root_.(string)
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.Mode().IsRegular() {
					return nil
				}

				select {
				case p.Output <- path:
				case <-p.Quit:
					return fmt.Errorf("done")
				}

				return nil
			})
			if err != nil {
				//log.Fatal(err)
			}
		case <-p.Quit:
			return
		}
	}
}

// A result is the product of reading and summing a file using MD5.
type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

// digester reads path names from paths and sends digests of the corresponding
// files on c until either paths or done is closed.
// This will be refered as the second stage
func digester(p Pipes) {
	defer p.Done.Done()

	for {
		select {
		case input := <-p.Input:
			if input == nil {
				return
			}
			path := input.(string)
			data, err := ioutil.ReadFile(path)
			select {
			case p.Output <- result{path, md5.Sum(data), err}:
			case <-p.Quit:
				return
			}
		case <-p.Quit:
			return
		}
	}
}

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error.  In that case,
// MD5All does not wait for inflight read operations to complete.
func MD5All(roots []string) (map[string][md5.Size]byte, error) {
	// MD5All closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	numBackend := runtime.NumCPU()

	// These channels link the different stages together:
	inputChan := make(chan interface{}, numBackend*2)
	fileChan := make(chan interface{}, numBackend*2)
	outputChan := make(chan interface{})
	quitChan := make(chan struct{})

	errChan := make(chan error)

	// Control number of spawned goroutines here:
	firstWg := sync.WaitGroup{}
	secondWg := sync.WaitGroup{}
	lastWg := sync.WaitGroup{}

	// These define how the stages connect to each stage
	// Note that the input for pipe2 is the output of pipe1
	pipe1 := Pipes{
		Input:  inputChan,
		Output: fileChan,
		Err:    errChan,
		Quit:   quitChan,
		Done:   &firstWg,
	}
	pipe2 := Pipes{
		Input:  fileChan,
		Output: outputChan,
		Err:    errChan,
		Quit:   quitChan,
		Done:   &secondWg,
	}
	// More concisely, the last stage:
	pipe3 := Pipes{outputChan, nil, errChan, quitChan, &lastWg}

	for i := 0; i < numBackend; i++ {
		// 1st stage
		firstWg.Add(1)
		go walkFiles(pipe1)
	}
	for i := 0; i < numBackend; i++ {
		// 2nd stage
		secondWg.Add(1)
		go digester(pipe2)
	}

	// We don't really stream the results, just collect them in
	// final stage in one map:
	m := make(map[string][md5.Size]byte)

	// Final stage, the "reduce" of this pipeline. One goroutine only
	lastWg.Add(1) // Only one goroutine!
	go func(p Pipes) {
		defer p.Done.Done()

		for {
			select {
			case r := <-p.Input:
				if r == nil {
					return
				}
				if r.(result).err != nil {
					return
				}
				m[r.(result).path] = r.(result).sum
			case <-p.Quit:
				return
			}
		}
	}(pipe3)

	// All set up, so send the input down the pipes we set up:
	for _, root := range roots {
		inputChan <- root
	}

	// This is the test part to force all goroutines to gracefully quit
	// after 3 seconds: this simulates a deadline:
	go func() {
		time.Sleep(time.Second * 3)
		for i := 0; i < numBackend*2+1; i++ {
			quitChan <- struct{}{}
		}
	}()

	// Now we're done input:
	close(inputChan)
	// Wait for first pipes to clear:
	firstWg.Wait()
	close(fileChan)
	// Wait for 2nd stage pipes to clear:
	secondWg.Wait()
	close(outputChan)
	// Wait for final pipe to clear:
	lastWg.Wait()

	return m, nil
}

func main() {
	go func() {
		for {
			log.Println("Backends ", runtime.NumGoroutine())
			time.Sleep(time.Second)
		}
	}()
	// Calculate the MD5 sum of all files under the specified directory,
	// then print the results sorted by path name.
	m, err := MD5All(os.Args[1:])
	if err != nil {
		fmt.Println("err!!!", err)
		return
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%x  %s\n", m[path], path)
	}
}

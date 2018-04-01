package main

import (
	"fmt"
)

// https://garbagecollected.org/2017/02/22/go-range-loop-internals/
// Today's #golang gotcha: the two-value range over an array does a copy. Avoid by ranging over the pointer instead
func IndexValueArray() {
	a := [...]int{1, 2, 3, 4, 5, 6, 7, 8}

	for i, v := range a {
		a[3] = 100 // Overwrite a[3] all the time (so even before the loop reaches 3).
		if i == 3 {
			fmt.Println("IndexValueArray", i, v) // Still prints 4 as range made a copy of 'a'.
		}
	}
}

func IndexValueArrayPtr() {
	a := [...]int{1, 2, 3, 4, 5, 6, 7, 8}

	for i, v := range &a { // Use a pointer to 'a' so changes to a inside the loop are reflected to the loop values.
		a[3] = 100
		if i == 3 {
			fmt.Println("IndexValueArrayPtr", i, v)
		}
	}
}

func main() {
	IndexValueArray()
	IndexValueArrayPtr()
}

package main

import (
	"fmt"
	"log"
)

func LogRequest(request []byte) {
	log.Println("Request is")
	fmt.Println("====")
	fmt.Println(string(request))
	fmt.Println("====")
	log.Println("")
}

func StringInSlice(searchFor string, in []string) bool {
	for _, currentValue := range in {
		if currentValue == searchFor {
			return true
		}
	}
	return false
}

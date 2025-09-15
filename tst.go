package main

import "fmt"

func main() {
	allowedMethods := map[string]struct{}{
		"GET":  {},
		"POST": {},
	}
	_, ok := allowedMethods["GET"]
	fmt.Println(ok)
}

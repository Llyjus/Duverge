package main

import {
	"fmt",
	"net/http",
}

func main() {
	fmt.Println("Gateway is running on port 26526...")


	err := http.ListenAndServe(":26526", nil)
	if err != nil {
		panic(err)
	}
}
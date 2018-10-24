package main

import (
	"flag"
	"fmt"
)

func main() {
	api := flag.String("api", Github, "github or typicode")
	q := flag.String("q", "", "github login or typicode user id")
	flag.Parse()

	user, err := FetchUserByID(*api, *q)
	if err != nil {
		fmt.Printf("\u279c  %s\n", err.Error())
		return
	}
	fmt.Printf("Result \u279c  %+v\n", user)
}

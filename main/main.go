package main

import (
	"totoro"
)

func main() {
	ch := make(chan struct{})
	ttr := totoro.MakeTotoro()
	ttr.Start()
	for range ch{

	}
}

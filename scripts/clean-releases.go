package main

import (
	"fmt"
	"os"
	"os/exec"
)

func task(i int) {
	cmd := exec.Command("sh", "./scripts/unrelease-version.sh", fmt.Sprintf("v0.0.%d", i))
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Done: ", i)
	ch <- 0
}

var ch = make(chan byte, 100)

func main() {
	for i := 0; i < 100; i++ {
		go task(i)
	}

	for i := 0; i < 100; i++ {
		<-ch
	}
}

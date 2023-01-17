package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	log.Println("Logs from your program will appear here!")

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

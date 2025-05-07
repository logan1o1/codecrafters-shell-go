package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

var builtins = []string{"exit", "echo", "type"}

func main() {

	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input: ", err)
			os.Exit(1)
		}
		cmdArr := strings.Fields(input)
		cmd := cmdArr[0]

		if strings.TrimSpace(input) == "exit 0" {
			break
		}

		switch cmd {
		case "echo":
			EchoCmd(cmdArr)
		case "type":
			TypeCmd(cmdArr)
		default:
			fmt.Println(strings.TrimSpace(cmd) + ": command not found")
		}

	}

}

func EchoCmd(cmdArr []string) {
	output := strings.Join(cmdArr[1:], " ")
	fmt.Println(output)
}

func TypeCmd(cmdArr []string) {
	if len(cmdArr) == 1 {
		return
	}

	value := cmdArr[1]

	if slices.Contains(builtins, value) {
		fmt.Println(value + " is a shell builtin")
		return
	}

	if filePath, exists := FindPath(value); exists {
		fmt.Println(value + " is " + filePath)
		return
	}

	fmt.Println(value + ": not found")
}

func FindPath(val string) (string, bool) {
	pathValue := os.Getenv("PATH")
	for _, path := range strings.Split(pathValue, ":") {
		file := path + "/" + val
		if _, err := os.Stat(file); err == nil {
			return file, true
		}
	}
	return "", false
}

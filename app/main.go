package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var builtins = []string{"exit", "echo", "type"}

func main() {

	for {
		fmt.Fprint(os.Stdout, "$ ")

		paths := strings.Split(os.Getenv("PATH"), ":")
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
			TypeCmd(cmdArr, paths)
		default:
			fmt.Println(strings.TrimSpace(cmd) + ": command not found")
		}
	}

}

func EchoCmd(cmdArr []string) {
	output := strings.Join(cmdArr[1:], " ")
	fmt.Println(output)
}

func TypeCmd(cmdArr []string, paths []string) {
	if len(cmdArr) == 1 {
		return
	}

	value := cmdArr[1]

	if slices.Contains(builtins, value) {
		fmt.Println(value + " is a shell builtin")
		return
	}

	if filePath, exists := FindPath(value, paths); exists {
		fmt.Println(value + " is " + filePath)
		return
	}

	fmt.Println(value + ": not found")
}

func CustomExeCmd(cmdArr []string, paths []string) {
	if len(cmdArr) == 1 {
		return
	}

	value := cmdArr[1]
	programName := cmdArr[0]
	numOfArg := len(cmdArr)
	programSign := rand.Uint64()

	if _, exists := FindPath(programName, paths); exists {
		fmt.Printf("Program was passed %d args (including program name).\n", numOfArg)
		fmt.Printf("Arg #0 (program name): %s\n", programName)
		fmt.Printf("Arg #1: %s\n", value)
		fmt.Printf("Program Signature: %d\n", programSign)
	}
}

func FindPath(val string, paths []string) (string, bool) {
	for _, path := range paths {
		file := path + "/" + val
		if _, err := os.Stat(file); err == nil {
			return file, true
		}
	}
	return "", false
}

func FindExecutables(cmd string, paths []string) string {
	for _, path := range paths {
		filepath := filepath.Join(path, cmd)
		fileinfo, err := os.Stat(filepath)
		if err != nil && fileinfo.Mode().Perm()&0111 != 0 {
			return filepath
		}
	}
	return ""
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
		cmd := strings.TrimSpace(cmdArr[0])
		args := cmdArr[1:]

		if strings.TrimSpace(input) == "exit 0" {
			break
		}

		switch cmd {
		case "echo":
			EchoCmd(cmdArr)
		case "type":
			TypeCmd(cmdArr, paths)
		default:
			filepath, exists := FindPath(cmd, paths)
			if exists && filepath != "" {
				CustomExeCmd(cmd, args)
			} else {
				fmt.Println(cmd + ": command not found")
			}
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

func CustomExeCmd(cmd string, args []string) {
	exc := exec.Command(cmd, args...)
	exc.Stdout = os.Stdout
	exc.Stderr = os.Stderr
	exc.Run()
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

// func FindExecutables(cmd string, paths []string) string {
// 	filepath, exists := FindPath(cmd, paths)
// 	_, err := os.Stat(filepath)
// 	if err != nil && exists {
// 		return filepath
// 	}
// 	return ""
// }

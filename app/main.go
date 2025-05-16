package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

var builtins = []string{"exit", "echo", "type", "pwd"}

func FindPath(val string, paths []string) (string, bool) {
	for _, path := range paths {
		file := path + "/" + val
		if _, err := os.Stat(file); err == nil {
			return file, true
		}
	}
	return "", false
}

func RemoveQuotes(input string) (string, []string) {
	var word string
	var output string
	var newArr []string
	var quoteChar rune
	// var slashChar rune
	if strings.Contains(input, "'") || strings.Contains(input, `"`) || strings.Contains(input, `/`) || strings.Contains(input, `\`) {

		for i := 0; i < len(input); i++ {
			ch := rune(input[i])

			switch ch {
			case '\'', '"':
				if quoteChar == 0 {
					quoteChar = ch
				} else if quoteChar == ch {
					if !(i+1 < len(input) && (input[i+1] == '\'' || input[i+1] == '"' || input[i+1] == '\\' || input[i+1] == '/')) {
						newArr = append(newArr, word)
						word = ""
					}
					quoteChar = 0
				} else {
					word += string(ch)
				}
			case '\\', '/':
				// I'm having problem with this case, I am not sure how to implement the 2nd part of the challenge here
			default:
				if quoteChar != 0 {
					word += string(ch)
				}
			}
		}

		if word != "" && quoteChar != 0 {
			newArr = append(newArr, word)
		}

		singlesRemoved := strings.ReplaceAll(input, "'", "")
		doublesRemoved := strings.ReplaceAll(singlesRemoved, `"`, "")
		output = doublesRemoved
	} else {
		forwardRemoved := strings.ReplaceAll(input, `/`, "")
		backwardRemoved := strings.ReplaceAll(forwardRemoved, `\`, "")
		output = backwardRemoved
		// output = input
	}
	return output, newArr
}

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

		var args []string

		newInput, newArgsArr := RemoveQuotes(input)

		if strings.Contains(input, "'") || strings.Contains(input, `"`) || strings.Contains(input, `/`) || strings.Contains(input, `\`) {
			input = newInput
			args = newArgsArr
			fmt.Println(args)
		} else if len(newArgsArr) == 0 {
			args = cmdArr[1:]
		} else {
			args = cmdArr[1:]
		}

		if strings.TrimSpace(input) == "exit 0" {
			break
		}

		switch cmd {
		case "echo":
			EchoCmd(args)
		case "type":
			TypeCmd(cmdArr, paths)
		case "pwd":
			Pwd()
		case "cd":
			Cd(args)
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

func EchoCmd(args []string) {
	output := strings.Join(args, " ")
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

func Pwd() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	fmt.Println(dir)
}

func Cd(args []string) {
	path := strings.Join(args, "")
	homepath := os.Getenv("HOME")
	formatedPath := strings.ReplaceAll(path, "~", homepath)
	err := os.Chdir(formatedPath)
	if err != nil {
		fmt.Println("cd: " + path + ": No such file or directory")
	}
}

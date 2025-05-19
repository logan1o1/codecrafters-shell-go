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

func InputParser(input string) (string, []string) {
	var word strings.Builder
	var newArr []string
	preserveNextLiteral := false
	backslashInQuotes := false
	inQuotes := false
	quoteChar := rune(0)

	for _, ch := range input {

		if preserveNextLiteral {
			word.WriteRune(ch)
			preserveNextLiteral = false
			continue
		}
		if backslashInQuotes {
			if ch == '$' || ch == '\\' || ch == '"' {
				word.WriteRune(ch)
			} else {
				word.WriteRune('\\')
				word.WriteRune(ch)
			}
			backslashInQuotes = false
			continue
		}

		switch {
		case ch == '"' || ch == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuotes = false
				quoteChar = rune(0)
			} else {
				word.WriteRune(ch)
			}
		case ch == '\\':
			if !inQuotes {
				preserveNextLiteral = true
			} else if quoteChar == '"' {
				backslashInQuotes = true
			} else {
				word.WriteRune(ch)
			}
		case ch == ' ':
			if inQuotes {
				word.WriteRune(ch)
			} else if word.Len() > 0 {
				newArr = append(newArr, word.String())
				word.Reset()
			}
		default:
			word.WriteRune(ch)
		}
	}

	if word.Len() > 0 {
		newArr = append(newArr, word.String())
	}
	if len(newArr) == 0 {
		return "", nil
	}

	noSingles := strings.ReplaceAll(input, "'", "")
	noDoubles := strings.ReplaceAll(noSingles, `"`, "")
	output := noDoubles

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

		newInput, newArgsArr := InputParser(input[:len(input)-1])

		if strings.Contains(input, "'") || strings.Contains(input, `"`) || strings.Contains(input, `/`) || strings.Contains(input, `\`) {
			input = newInput
			args = newArgsArr[1:]
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

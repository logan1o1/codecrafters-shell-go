package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
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

		if !inQuotes && quoteChar == 0 && ch == '/' {
			word.WriteRune(ch)
			continue
		}

		if preserveNextLiteral {
			word.WriteRune(ch)
			preserveNextLiteral = false
			continue
		}
		if backslashInQuotes {
			if ch == '$' || ch == '\\' || ch == '"' || ch == '`' {
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
			cmd = newArgsArr[0]
		} else if len(newArgsArr) == 0 {
			args = cmdArr[1:]
		} else {
			args = cmdArr[1:]
		}

		if strings.TrimSpace(input) == "exit 0" {
			break
		}

		var outfile *os.File
		var errfile *os.File
		inTwo := false

		for i, arg := range args {
			if (arg == ">" || arg == "1>" || arg == "2>") && i+1 < len(args) {
				outfile, err = os.Create(args[i+1])
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating file:", err)
					continue
				}
				errfile, err = os.Create(args[i+1])
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating file:", err)
					continue
				}
				if arg == "2>" {
					inTwo = true
				} else {
					inTwo = false
				}
				args = args[:i]
				break
			} else if (arg == ">>" || arg == "1>>" || arg == "2>>") && i+1 < len(args) {
				outfile, err = os.OpenFile(args[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating file:", err)
					continue
				}
				errfile, err = os.OpenFile(args[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating file:", err)
					continue
				}
				if arg == "2>>" {
					inTwo = true
				} else {
					inTwo = false
				}
				args = args[:i]
				break
			}
		}

		if outfile != nil || errfile != nil {
			var originalStderr *os.File
			var originalStdout *os.File
			defer outfile.Close()
			defer errfile.Close()
			if inTwo {
				originalStderr = os.Stderr
				os.Stderr = errfile
			} else {
				originalStdout = os.Stdout
				os.Stdout = outfile
			}
			defer func() {
				os.Stdout = originalStdout
				os.Stderr = originalStderr
			}()
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
				// ExecuteAndRedirect(cmd, args)
			} else {
				fmt.Println(cmd + ": command not found")
			}
		}

		if outfile != nil {
			os.Stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
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
	err := exc.Run()
	if err != nil {
		return
	}
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

func ExecuteAndRedirect(cmd string, args []string) {
	if cmd == "" || len(args) == 0 {
		return
	}

	var outfile *os.File
	var err error

	filtered := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		token := args[i]
		if (token == ">" || token == "1>") && i+1 < len(args) {
			outName := args[i+1]
			outfile, err = os.OpenFile(outName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
			if err != nil {
				fmt.Fprintln(os.Stderr, "redirect error:", err)
				return
			}
			i++
		} else {
			filtered = append(filtered, token)
		}
	}

	command := exec.Command(cmd, filtered...)
	if outfile != nil {
		command.Stdout = outfile
	} else {
		command.Stdout = os.Stdout
	}
	command.Stderr = os.Stderr

	if outfile != nil {
		outfile.Close()
	}
}

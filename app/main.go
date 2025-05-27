package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
)

var builtins = []string{"exit", "echo", "type", "pwd", "history"}

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

func CustomExeFromPath() []string {
	envPath := os.Getenv("PATH")
	paths := strings.Split(envPath, string(os.PathListSeparator))
	seen := make(map[string]bool)
	var result []string

	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			name := file.Name()
			if seen[name] {
				continue
			}
			fullPath := filepath.Join(dir, name)
			info, err := os.Stat(fullPath)
			if err == nil && !info.IsDir() && (info.Mode()&0111 != 0) {
				seen[name] = true
				result = append(result, name)
			}
		}
	}
	return result
}

func getCurrentWord(line []rune, pos int) string {
	start := pos

	for start > 0 && line[start-1] != ' ' {
		start--
	}
	end := pos

	for end < len(line) && line[end] != ' ' {
		end++
	}

	return string(line[start:end])
}

func printMatchesInline(cmds []string, prefix string) {

	var matches []string

	for _, cmd := range cmds {

		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}

	}

	if len(matches) > 0 {
		fmt.Print("\n")

		for _, match := range matches {
			fmt.Print(match, "  ")
		}
		fmt.Print("\n")
	}
}

type tabListener struct {
	lastTabTime time.Time

	cachedExecutables []string

	rl *readline.Instance
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, str := range strs[1:] {
		i := 0
		for i < len(prefix) && i < len(str) && prefix[i] == str[i] {
			i++
		}
		prefix = prefix[:i]
		if prefix == "" {
			break
		}
	}
	return prefix
}

func (l *tabListener) OnChange(line []rune, pos int, key rune) ([]rune, int, bool) {

	if key == readline.CharTab {

		now := time.Now()
		word := getCurrentWord(line, pos)

		// Collect matches
		var matches []string

		for _, cmd := range l.cachedExecutables {
			if strings.HasPrefix(cmd, word) {
				matches = append(matches, cmd)
			}
		}

		if len(matches) == 0 {
			fmt.Print("\x07") // bell - no matches
			l.lastTabTime = now
			return line, pos, true
		}

		if len(matches) == 1 {
			suffix := matches[0][len(word):]
			// Add trailing space after completion if suffix is empty or cursor is at the end
			if suffix == "" {
				suffix = " "
			} else {
				// You may want to add a space if the cursor is at the end of the word
				// but generally adding space after completion is safe
				suffix += " "
			}
			// Insert suffix at cursor position
			newLine := make([]rune, 0, len(line)+len(suffix))
			newLine = append(newLine, line[:pos]...)
			newLine = append(newLine, []rune(suffix)...)
			newLine = append(newLine, line[pos:]...)
			newPos := pos + len(suffix)
			l.lastTabTime = now

			return newLine, newPos, true
		}

		lcp := longestCommonPrefix(matches)
		if len(lcp) > len(word) {
			suffix := lcp[len(word):]
			newLine := make([]rune, 0, len(line)+len(suffix))
			newLine = append(newLine, line[:pos]...)
			newLine = append(newLine, []rune(suffix)...)
			newLine = append(newLine, line[pos:]...)
			newPos := pos + len(suffix)
			l.lastTabTime = now

			return newLine, newPos, true

		}
		// Multiple matches
		if now.Sub(l.lastTabTime) < 500*time.Millisecond {
			// Second tab — print matches
			fmt.Print("\n")
			for _, m := range matches {
				fmt.Print(m, "  ")
			}
			fmt.Print("\n")
			l.rl.Refresh()
		} else {
			// First tab — bell
			fmt.Print("\x07")
		}
		l.lastTabTime = now
		return line, pos, true
	}
	return line, pos, false
}

func (l *tabListener) OnExecute([]rune) {}

func main() {
	cachedExe := CustomExeFromPath()
	cachedExe = append(cachedExe, "exit", "type", "cd")
	sort.Strings(cachedExe)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "$ ",
		AutoComplete: readline.PcItemDynamic(func(s string) []string {
			return nil
		}),
	})

	if err != nil {
		panic(err)
	}

	defer rl.Close()

	rl.Config.Listener = &tabListener{
		cachedExecutables: cachedExe,
		rl:                rl,
	}

	var history []string

	for {
		fmt.Fprint(os.Stdout, "$ ")

		paths := strings.Split(os.Getenv("PATH"), ":")

		input, err := rl.Readline()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input: ", err)
			os.Exit(1)
		}

		cmdArr := strings.Fields(input)
		cmd := strings.TrimSpace(cmdArr[0])

		newInput, newArgsArr := InputParser(input)
		history = append(history, newInput)

		var args []string

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
		case "history":
			for i := range len(history) {
				fmt.Printf("%d  %s\n", i, history[i])
			}
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

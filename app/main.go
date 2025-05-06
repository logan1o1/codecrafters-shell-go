package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {

	for {
		fmt.Fprint(os.Stdout, "$ ")

		cmd, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input: ", err)
			os.Exit(1)
		}

		if strings.TrimSpace(cmd) == "exit 0" {
			break
		}

		if strings.Contains(cmd, "echo") {
			inputArr := strings.Fields(cmd)
			outputArr := inputArr[1:]
			output := strings.Join(outputArr, " ")

			fmt.Println(output)
		} else {
			fmt.Println(strings.TrimSpace(cmd) + ": command not found")
		}

	}

}

package cmd

import (
	"bufio"
	"os"
	"strings"
)

// ReadLine reads a line of text from standard input, trimming trailing newline characters.
func ReadLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	option, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	option = strings.TrimSpace(option)

	return option, nil
}


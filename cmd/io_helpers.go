package cmd

import (
	"bufio"
	"errors"
	"gymrat/models"
	"os"
	"strconv"
	"strings"
)

var maxUnitValues = models.MaxUnitValues // var to map constants

// read input from user
func ReadLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	//Read string until you see the user hit enter
	option, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	option = strings.TrimSpace(option) // trim the \n

	return option, nil
}

func GetMeAValidMaxValue(str string, isZeroAllowed bool, maxAllowed int) (int, error) {

	defaultZero := 0
	newIntValue, err := strconv.Atoi(str)
	if err != nil {
		return defaultZero, err
	}

	if !isZeroAllowed && newIntValue == 0 {
		return defaultZero, errors.New("error: zero is not allowed")
	}

	if newIntValue > maxAllowed {
		return defaultZero, errors.New("error: value is over the max allowed")
	}

	return newIntValue, nil
}

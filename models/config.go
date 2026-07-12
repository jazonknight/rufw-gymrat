package models

import (
	"errors"
	"strconv"
)

// define map of max units
var MaxUnitValues = map[UnitType]int{
	"seconds": 360,
	"count":   30,
	"sets":    5,
	"effort":  10,
}

type FormationType string
type UnitType string

const (
	BothSides FormationType = "BOTH_SIDES"
	LeftSide  FormationType = "LEFT_SIDE"
	RightSide FormationType = "RIGHT_SIDE"
	Time      UnitType      = "seconds"
	Sequence  UnitType      = "count"
)

func IsValidFormation(input string) bool {
	switch FormationType(input) {
	case BothSides, LeftSide, RightSide:
		return true
	default:
		return false
	}
}

func IsValidUnit(input string) bool {
	switch UnitType(input) {
	case Time, Sequence:
		return true
	default:
		return false
	}
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

func GetMaxPerUnit(unitStr UnitType) (int, error) {
	//maxValue := 0

	/*
		switch str {
		case "seconds":
			maxValue = 360
		case "count":
			maxValue = 30
		default:
			return 0, errors.New("error: invalid unit provided")
		}
	*/

	if maxValue, exists := MaxUnitValues[unitStr]; exists {
		return maxValue, nil
	}

	return 0, errors.New("error: invalid unit provided")

}

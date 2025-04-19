package services

import (
	"errors"
	"strconv"
	"time"
)

const (
	// TR-106 parameter types
	ParameterType_base64       = "base64"
	ParameterType_boolean      = "boolean"
	ParameterType_dateTime     = "dateTime"
	ParameterType_hexBinary    = "hexBinary"
	ParameterType_int          = "int"
	ParameterType_long         = "long"
	ParameterType_unsignedInt  = "unsignedInt"
	ParameterType_unsignedLong = "unsignedLong"
	ParameterType_string       = "string"
	// TR-106 parameter types
)

var (
	ErrInvalidParameterType = errors.New("invalid parameter type")
)

func IsValidParameterType(parameterType string) bool {
	switch parameterType {
	case
		ParameterType_base64,
		ParameterType_boolean,
		ParameterType_dateTime,
		ParameterType_hexBinary,
		ParameterType_int,
		ParameterType_long,
		ParameterType_unsignedInt,
		ParameterType_unsignedLong,
		ParameterType_string:
		return true
	default:
		return false
	}
}

func ParseParameterValue(parameterType string, parameterValue string) (any, error) {
	var (
		value any
		err   error
	)

	switch parameterType {
	case ParameterType_base64:
		value, err = parameterValue, nil
	case ParameterType_boolean:
		value, err = strconv.ParseBool(parameterValue)
	case ParameterType_dateTime:
		value, err = time.Parse(time.RFC3339, parameterValue)
	case ParameterType_hexBinary:
		value, err = parameterValue, nil
	case ParameterType_int:
		value, err = strconv.ParseInt(parameterValue, 10, 32)
	case ParameterType_long:
		value, err = strconv.ParseInt(parameterValue, 10, 64)
	case ParameterType_unsignedInt:
		value, err = strconv.ParseUint(parameterValue, 10, 32)
	case ParameterType_unsignedLong:
		value, err = strconv.ParseUint(parameterValue, 10, 64)
	case ParameterType_string:
		value, err = parameterValue, nil
	default:
		value, err = nil, ErrInvalidParameterType
	}

	return value, err
}

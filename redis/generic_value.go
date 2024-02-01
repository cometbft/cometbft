package redis

import (
	"encoding/json"
	"strconv"
)

// GenericValue provides convenience methods for transforming between different types
type GenericValue struct {
	Raw string `json:"raw"`
}

// IsEmpty checks if the value is empty
func (generic GenericValue) IsEmpty() bool {
	return len(generic.Raw) == 0
}

// Present checks if the value is not empty
func (generic GenericValue) Present() bool {
	return len(generic.Raw) > 0
}

// String Returns GenericValue as a string
func (generic GenericValue) String() string {
	return generic.Raw
}

// StringArray Returns GenericValue as a string array
func (generic GenericValue) StringArray() []string {
	var array []string
	err := json.Unmarshal([]byte(generic.Raw), &array)
	if err != nil {
		panic(err)
	}
	return array
}

// Uint64 Returns GenericValue as a uint64
func (generic GenericValue) Uint64() uint64 {
	if generic.IsEmpty() {
		panic("Invalid Uint64")
	}
	parsedValue, err := strconv.ParseUint(generic.Raw, 10, 64)
	if err != nil {
		panic(err)
	}

	return parsedValue
}

// Int64 Returns GenericValue as a int64
func (generic GenericValue) Int64() int64 {
	if generic.IsEmpty() {
		panic("Invalid Int64")
	}
	parsedValue, err := strconv.ParseInt(generic.Raw, 10, 64)
	if err != nil {
		panic(err)
	}

	return parsedValue
}

// Float64 Returns GenericValue as a uint64
func (generic GenericValue) Float64() float64 {
	if generic.IsEmpty() {
		panic("Invalid Float64")
	}
	parsedValue, err := strconv.ParseFloat(generic.Raw, 64)
	if err != nil {
		panic(err)
	}

	return parsedValue
}

// Float64Array Returns GenericValue as a float64 array
func (generic GenericValue) Float64Array() []float64 {
	var array []float64
	err := json.Unmarshal([]byte(generic.Raw), &array)
	if err != nil {
		panic(err)
	}
	return array
}

// StringToGenericValue converts a string to a generic value
func StringToGenericValue(value string) GenericValue {
	return GenericValue{
		Raw: value,
	}
}

// Uint64ToGenericValue converts a uint64 to a generic value
func Uint64ToGenericValue(value uint64) GenericValue {
	return GenericValue{
		Raw: strconv.FormatUint(value, 10),
	}
}

// Int64ToGenericValue converts an int64 to a generic value
func Int64ToGenericValue(value int64) GenericValue {
	return GenericValue{
		Raw: strconv.FormatInt(value, 10),
	}
}

// Float64ToGenericValue converts a float64 to a generic value
func Float64ToGenericValue(value float64) GenericValue {
	return GenericValue{
		Raw: strconv.FormatFloat(value, 'f', -1, 64),
	}
}

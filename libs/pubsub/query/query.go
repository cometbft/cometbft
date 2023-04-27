// Package query provides a parser for a custom query format:
//
//	abci.invoice.number=22 AND abci.invoice.owner=Ivan
//
// See query.peg for the grammar, which is a https://en.wikipedia.org/wiki/Parsing_expression_grammar.
// More: https://github.com/PhilippeSigaud/Pegged/wiki/PEG-Basics
//
// It has a support for numbers (integer and floating point), dates and times.
package query

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	numRegex = regexp.MustCompile(`([0-9\.]+)`)
)

// Query holds the query string and the query parser.
type Query struct {
	str    string
	parser *QueryParser
}

// Condition represents a single condition within a query and consists of composite key
// (e.g. "tx.gas"), operator (e.g. "=") and operand (e.g. "7").
type Condition struct {
	CompositeKey string
	Op           Operator
	Operand      interface{}
}

// New parses the given string and returns a query or error if the string is
// invalid.
func New(s string) (*Query, error) {
	p := &QueryParser{Buffer: fmt.Sprintf(`"%s"`, s)}
	if err := p.Init(); err != nil {
		return nil, err
	}
	if err := p.Parse(); err != nil {
		return nil, err
	}
	return &Query{str: s, parser: p}, nil
}

// MustParse turns the given string into a query or panics; for tests or others
// cases where you know the string is valid.
func MustParse(s string) *Query {
	q, err := New(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s: %v", s, err))
	}
	return q
}

// String returns the original string.
func (q *Query) String() string {
	return q.str
}

// Operator is an operator that defines some kind of relation between composite key and
// operand (equality, etc.).
type Operator uint8

const (
	// "<="
	OpLessEqual Operator = iota
	// ">="
	OpGreaterEqual
	// "<"
	OpLess
	// ">"
	OpGreater
	// "="
	OpEqual
	// "CONTAINS"; used to check if a string contains a certain sub string.
	OpContains
	// "EXISTS"; used to check if a certain event attribute is present.
	OpExists
)

const (
	// DateLayout defines a layout for all dates (`DATE date`)
	DateLayout = "2006-01-02"
	// TimeLayout defines a layout for all times (`TIME time`)
	TimeLayout = time.RFC3339
)

// Conditions returns a list of conditions. It returns an error if there is any
// error with the provided grammar in the Query.
func (q *Query) Conditions() ([]Condition, error) {
	var (
		eventAttr string
		op        Operator
	)

	conditions := make([]Condition, 0)
	buffer, begin, end := q.parser.Buffer, 0, 0

	// tokens must be in the following order: tag ("tx.gas") -> operator ("=") -> operand ("7")
	for _, token := range q.parser.Tokens() {
		switch token.pegRule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)

		case ruletag:
			eventAttr = buffer[begin:end]

		case rulele:
			op = OpLessEqual

		case rulege:
			op = OpGreaterEqual

		case rulel:
			op = OpLess

		case ruleg:
			op = OpGreater

		case ruleequal:
			op = OpEqual

		case rulecontains:
			op = OpContains

		case ruleexists:
			op = OpExists
			conditions = append(conditions, Condition{eventAttr, op, nil})

		case rulevalue:
			// strip single quotes from value (i.e. "'NewBlock'" -> "NewBlock")
			valueWithoutSingleQuotes := buffer[begin+1 : end-1]
			conditions = append(conditions, Condition{eventAttr, op, valueWithoutSingleQuotes})

		case rulenumber:
			number := buffer[begin:end]
			if strings.ContainsAny(number, ".") { // if it looks like a floating-point number
				value, err := strconv.ParseFloat(number, 64)
				if err != nil {
					err = fmt.Errorf(
						"got %v while trying to parse %s as float64 (should never happen if the grammar is correct)",
						err, number,
					)
					return nil, err
				}

				conditions = append(conditions, Condition{eventAttr, op, value})
			} else {
				valueBig := new(big.Int)

				valueBig, ok := valueBig.SetString(number, 10)
				if !ok {
					err := fmt.Errorf(
						"problem parsing %s as bigint (should never happen if the grammar is correct)",
						number,
					)
					return nil, err
				}

				if valueBig.IsInt64() {
					conditions = append(conditions, Condition{eventAttr, op, valueBig.Int64()})
				} else {
					conditions = append(conditions, Condition{eventAttr, op, valueBig})
				}
			}

		case ruletime:
			value, err := time.Parse(TimeLayout, buffer[begin:end])
			if err != nil {
				err = fmt.Errorf(
					"got %v while trying to parse %s as time.Time / RFC3339 (should never happen if the grammar is correct)",
					err, buffer[begin:end],
				)
				return nil, err
			}

			conditions = append(conditions, Condition{eventAttr, op, value})

		case ruledate:
			value, err := time.Parse("2006-01-02", buffer[begin:end])
			if err != nil {
				err = fmt.Errorf(
					"got %v while trying to parse %s as time.Time / '2006-01-02' (should never happen if the grammar is correct)",
					err, buffer[begin:end],
				)
				return nil, err
			}

			conditions = append(conditions, Condition{eventAttr, op, value})
		}
	}

	return conditions, nil
}

// Matches returns true if the query matches against any event in the given set
// of events, false otherwise. For each event, a match exists if the query is
// matched against *any* value in a slice of values. An error is returned if
// any attempted event match returns an error.
//
// For example, query "name=John" matches events = {"name": ["John", "Eric"]}.
// More examples could be found in parser_test.go and query_test.go.
func (q *Query) Matches(events map[string][]string) (bool, error) {
	if len(events) == 0 {
		return false, nil
	}

	var (
		eventAttr string
		op        Operator
	)

	buffer, begin, end := q.parser.Buffer, 0, 0

	// tokens must be in the following order:

	// tag ("tx.gas") -> operator ("=") -> operand ("7")
	for _, token := range q.parser.Tokens() {
		switch token.pegRule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)

		case ruletag:
			eventAttr = buffer[begin:end]

		case rulele:
			op = OpLessEqual

		case rulege:
			op = OpGreaterEqual

		case rulel:
			op = OpLess

		case ruleg:
			op = OpGreater

		case ruleequal:
			op = OpEqual

		case rulecontains:
			op = OpContains
		case ruleexists:
			op = OpExists
			if strings.Contains(eventAttr, ".") {
				// Searching for a full "type.attribute" event.
				_, ok := events[eventAttr]
				if !ok {
					return false, nil
				}
			} else {
				foundEvent := false

			loop:
				for compositeKey := range events {
					if strings.Index(compositeKey, eventAttr) == 0 {
						foundEvent = true
						break loop
					}
				}
				if !foundEvent {
					return false, nil
				}
			}

		case rulevalue:
			// strip single quotes from value (i.e. "'NewBlock'" -> "NewBlock")
			valueWithoutSingleQuotes := buffer[begin+1 : end-1]

			// see if the triplet (event attribute, operator, operand) matches any event
			// "tx.gas", "=", "7", { "tx.gas": 7, "tx.ID": "4AE393495334" }
			match, err := match(eventAttr, op, reflect.ValueOf(valueWithoutSingleQuotes), events)
			if err != nil {
				return false, err
			}

			if !match {
				return false, nil
			}

		case rulenumber:
			number := buffer[begin:end]
			if strings.ContainsAny(number, ".") { // if it looks like a floating-point number
				value, err := strconv.ParseFloat(number, 64)
				if err != nil {
					err = fmt.Errorf(
						"got %v while trying to parse %s as float64 (should never happen if the grammar is correct)",
						err, number,
					)
					return false, err
				}

				match, err := match(eventAttr, op, reflect.ValueOf(value), events)
				if err != nil {
					return false, err
				}

				if !match {
					return false, nil
				}
			} else {
				value := new(big.Int)
				_, ok := value.SetString(number, 10)

				if !ok {
					err := fmt.Errorf(
						"problem parsing %s as bigInt (should never happen if the grammar is correct)",
						number,
					)
					return false, err
				}

				match, err := match(eventAttr, op, reflect.ValueOf(value), events)
				if err != nil {
					return false, err
				}

				if !match {
					return false, nil
				}
			}

		case ruletime:
			value, err := time.Parse(TimeLayout, buffer[begin:end])
			if err != nil {
				err = fmt.Errorf(
					"got %v while trying to parse %s as time.Time / RFC3339 (should never happen if the grammar is correct)",
					err, buffer[begin:end],
				)
				return false, err
			}

			match, err := match(eventAttr, op, reflect.ValueOf(value), events)
			if err != nil {
				return false, err
			}

			if !match {
				return false, nil
			}

		case ruledate:
			value, err := time.Parse("2006-01-02", buffer[begin:end])
			if err != nil {
				err = fmt.Errorf(
					"got %v while trying to parse %s as time.Time / '2006-01-02' (should never happen if the grammar is correct)",
					err, buffer[begin:end],
				)
				return false, err
			}

			match, err := match(eventAttr, op, reflect.ValueOf(value), events)
			if err != nil {
				return false, err
			}

			if !match {
				return false, nil
			}
		}
	}

	return true, nil
}

// match returns true if the given triplet (attribute, operator, operand) matches
// any value in an event for that attribute. If any match fails with an error,
// that error is returned.
//
// First, it looks up the key in the events and if it finds one, tries to compare
// all the values from it to the operand using the operator.
//
// "tx.gas", "=", "7", {"tx": [{"gas": 7, "ID": "4AE393495334"}]}
func match(attr string, op Operator, operand reflect.Value, events map[string][]string) (bool, error) {
	// look up the tag from the query in tags
	values, ok := events[attr]
	if !ok {
		return false, nil
	}

	for _, value := range values {
		// return true if any value in the set of the event's values matches
		match, err := matchValue(value, op, operand)
		if err != nil {
			return false, err
		}

		if match {
			return true, nil
		}
	}

	return false, nil
}

// matchValue will attempt to match a string value against an operator an
// operand. A boolean is returned representing the match result. It will return
// an error if the value cannot be parsed and matched against the operand type.
func matchValue(value string, op Operator, operand reflect.Value) (bool, error) {
	switch operand.Kind() {
	case reflect.Struct: // time
		operandAsTime := operand.Interface().(time.Time)

		// try our best to convert value from events to time.Time
		var (
			v   time.Time
			err error
		)

		if strings.ContainsAny(value, "T") {
			v, err = time.Parse(TimeLayout, value)
		} else {
			v, err = time.Parse(DateLayout, value)
		}
		if err != nil {
			return false, fmt.Errorf("failed to convert value %v from event attribute to time.Time: %w", value, err)
		}

		switch op {
		case OpLessEqual:
			return (v.Before(operandAsTime) || v.Equal(operandAsTime)), nil
		case OpGreaterEqual:
			return (v.Equal(operandAsTime) || v.After(operandAsTime)), nil
		case OpLess:
			return v.Before(operandAsTime), nil
		case OpGreater:
			return v.After(operandAsTime), nil
		case OpEqual:
			return v.Equal(operandAsTime), nil
		}

	case reflect.Float64:
		var v float64

		operandFloat64 := operand.Interface().(float64)
		filteredValue := numRegex.FindString(value)

		// try our best to convert value from tags to float64
		v, err := strconv.ParseFloat(filteredValue, 64)
		if err != nil {
			return false, fmt.Errorf("failed to convert value %v from event attribute to float64: %w", filteredValue, err)
		}

		switch op {
		case OpLessEqual:
			return v <= operandFloat64, nil
		case OpGreaterEqual:
			return v >= operandFloat64, nil
		case OpLess:
			return v < operandFloat64, nil
		case OpGreater:
			return v > operandFloat64, nil
		case OpEqual:
			return v == operandFloat64, nil
		}

	case reflect.Pointer:

		var i *big.Int
		if reflect.TypeOf(operand.Interface()) != reflect.TypeOf(i) {
			break
		}

		filteredValue := numRegex.FindString(value)
		var cmpRes int
		if strings.ContainsAny(filteredValue, ".") {
			floatVal := new(big.Float)
			_, ok := floatVal.SetString(operand.Interface().(*big.Int).String())
			if !ok {
				return false, fmt.Errorf("failed to convert value %v from event attribute to float64", filteredValue)
			}
			v := new(big.Float)
			v, ok = v.SetString(filteredValue)
			if !ok {
				return false, fmt.Errorf("failed to convert value %v from event attribute to float64", filteredValue)
			}
			cmpRes = floatVal.Cmp(v)
		} else {
			operandVal := operand.Interface().(*big.Int)
			// try our best to convert value from tags to int64
			v := new(big.Int)

			v, ok := v.SetString(filteredValue, 10)

			if !ok {
				return false, fmt.Errorf("failed to convert value %v from event attribute to big int", filteredValue)
			}
			cmpRes = operandVal.Cmp(v)
		}

		switch op {
		case OpLessEqual:
			return cmpRes == 0 || cmpRes == 1, nil
		case OpGreaterEqual:
			return cmpRes == 0 || cmpRes == -1, nil
		case OpLess:
			return cmpRes == 1, nil
		case OpGreater:
			return cmpRes == -1, nil
		case OpEqual:
			return cmpRes == 0, nil
		}

	case reflect.Int64:
		var v int64

		operandInt := operand.Interface().(int64)
		filteredValue := numRegex.FindString(value)
		noFrac := true
		// if value looks like float, we try to parse it as float
		if strings.ContainsAny(filteredValue, ".") {
			v1, err := strconv.ParseFloat(filteredValue, 64)
			if err != nil {
				return false, fmt.Errorf("failed to convert value %v from event attribute to float64: %w", filteredValue, err)
			}
			if _, frac := math.Modf(v1); frac != 0 {
				noFrac = false // the numbers cannot be equal if the floating point has anything else than 0 as fraction
			}
			v = int64(v1)
		} else {
			var err error
			// try our best to convert value from tags to int64
			v, err = strconv.ParseInt(filteredValue, 10, 64)
			if err != nil {
				return false, fmt.Errorf("failed to convert value %v from event attribute to int64: %w", filteredValue, err)
			}
		}

		switch op {
		case OpLessEqual:
			return v < operandInt || (v == operandInt && noFrac), nil
		case OpGreaterEqual:
			return v > operandInt || (v == operandInt && noFrac), nil
		case OpLess:
			return v < operandInt, nil
		case OpGreater:
			// If the float had fractions that were removed by the cast in L527 we need to check the second part of the condition
			return v > operandInt || (v == operandInt && !noFrac), nil
		case OpEqual:
			// noFrac confirms that they are actually equal in value
			return v == operandInt && noFrac, nil
		}

	case reflect.String:
		switch op {
		case OpEqual:
			return value == operand.String(), nil
		case OpContains:
			return strings.Contains(value, operand.String()), nil
		}

	default:
		return false, fmt.Errorf("unknown kind of operand %v", operand.Kind())
	}

	return false, nil
}

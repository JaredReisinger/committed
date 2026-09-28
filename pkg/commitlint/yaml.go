package commitlint

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// obnoxious these aren't exported from yaml package
const (
	nullTag      = "!!null"
	boolTag      = "!!bool"
	strTag       = "!!str"
	intTag       = "!!int"
	floatTag     = "!!float"
	timestampTag = "!!timestamp"
	seqTag       = "!!seq"
	mapTag       = "!!map"
	binaryTag    = "!!binary"
	mergeTag     = "!!merge"
)

func (v *RuleValue[valType]) UnmarshalYAML(node *yaml.Node) error {
	// wish I could include the parent name!
	slog.Debug("parsing rule values", "type", "YAML")

	noValue := reflect.TypeOf(v.Value) == reflect.TypeOf(NoValue{})
	expectedLen := 3
	if noValue {
		expectedLen--
	}

	if node.Kind != yaml.SequenceNode {
		return errors.New("expected !!seq")
	}

	if len(node.Content) != expectedLen {
		return fmt.Errorf("expected %d values in rule", expectedLen)
	}

	if node.Content[0].Tag != intTag {
		return errors.New("first value should be int")
	}

	n, err := strconv.Atoi(node.Content[0].Value)
	if err != nil {
		return err
	}

	v.Level = Level(n)

	if node.Content[1].Tag != strTag {
		return errors.New("second value should be string")
	}

	err = (&v.When).UnmarshalString(node.Content[1].Value)
	if err != nil {
		return err
	}

	// TODO: 3rd value could be anything...
	if !noValue {
		err = node.Content[2].Decode(&v.Value)
		if err != nil {
			return err
		}
	}

	v.Set = true

	slog.Debug("parsed rule values", "values", v)

	return nil
}

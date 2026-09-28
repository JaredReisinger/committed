package commitlint

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"reflect"
)

func (v *RuleValue[valType]) UnmarshalJSON(data []byte) error {
	// var tmp []any
	slog.Debug("parsing rule values", "type", "JSON", "data", string(data))

	noValue := reflect.TypeOf(v.Value) == reflect.TypeOf(NoValue{})
	expectedLen := 3
	if noValue {
		expectedLen--
	}

	// incoming is an array...
	var tmp []jsontext.Value

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	if len(tmp) != expectedLen {
		return fmt.Errorf("expected %d values in rule", expectedLen)
	}

	err = json.Unmarshal(tmp[0], &v.Level)
	if err != nil {
		return err
	}

	err = json.Unmarshal(tmp[1], &v.When)
	if err != nil {
		return err
	}

	if !noValue {
		err = json.Unmarshal(tmp[2], &v.Value)
		if err != nil {
			return err
		}
	}

	v.Set = true
	slog.Debug("parsed rule values", "values", v)

	return nil
}

package commitlint

import (
	"encoding/json"
	"errors"
	"log/slog"
)

func (v *RuleValue[valType]) UnmarshalJSON(data []byte) error {
	// var tmp []any
	slog.Debug("parsing rule values", "type", "JSON", "data", string(data))

	// incoming is an array...
	var tmp []json.RawMessage

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	if len(tmp) != 3 {
		return errors.New("expected three values in rule")
	}

	err = json.Unmarshal(tmp[0], &v.Level)
	if err != nil {
		return err
	}

	err = json.Unmarshal(tmp[1], &v.When)
	if err != nil {
		return err
	}

	err = json.Unmarshal(tmp[2], &v.Value)
	if err != nil {
		return err
	}

	v.Set = true
	slog.Debug("parsed rule values", "values", v)

	return nil
}

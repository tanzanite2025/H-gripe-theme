package market

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringList []string

func (l StringList) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]string(l))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (l *StringList) Scan(value interface{}) error {
	if l == nil {
		return nil
	}
	if value == nil {
		*l = StringList{}
		return nil
	}

	var data []byte
	switch typed := value.(type) {
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	default:
		return fmt.Errorf("scan string list: unsupported value type %T", value)
	}

	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	*l = StringList(values)
	return nil
}

func (l StringList) Slice() []string {
	return append([]string(nil), []string(l)...)
}

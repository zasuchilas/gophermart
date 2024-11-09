package converters

import (
	"errors"
	"strconv"
)

func InterfaceToInt64(i interface{}) (int64, error) {
	switch v := i.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, errors.New("type error")
	}
}

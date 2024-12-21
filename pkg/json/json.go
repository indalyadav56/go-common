package json

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Marshal converts any value to JSON bytes.
func Marshal(v interface{}) ([]byte, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal to JSON failed: %w", err)
	}
	return bytes, nil
}

// Unmarshal converts JSON bytes to a value.
func Unmarshal(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal from JSON failed: %w", err)
	}
	return nil
}

// MarshalToString converts any value to a JSON string.
func MarshalToString(v interface{}) (string, error) {
	bytes, err := Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// UnmarshalFromString converts a JSON string to a value.
func UnmarshalFromString(data string, v interface{}) error {
	return Unmarshal([]byte(data), v)
}

// PrettyPrint returns a pretty-printed JSON string.
func PrettyPrint(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return "", fmt.Errorf("pretty print JSON failed: %w", err)
	}
	return string(bytes), nil
}

// ValidateJSON checks if a JSON byte array is valid.
func ValidateJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// ValidateJSONString checks if a JSON string is valid.
func ValidateJSONString(data string) error {
	return ValidateJSON([]byte(data))
}

// DeepCopy makes a deep copy of a JSON-serializable value.
// The src parameter is the source value to copy from.
// The dest parameter is a pointer to the destination where the copy will be stored.
func DeepCopy(src, dest interface{}) error {
	bytes, err := Marshal(src)
	if err != nil {
		return fmt.Errorf("deep copy serialization failed: %w", err)
	}

	if err := Unmarshal(bytes, dest); err != nil {
		return fmt.Errorf("deep copy deserialization failed: %w", err)
	}
	return nil
}

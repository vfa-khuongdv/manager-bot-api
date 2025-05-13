package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// CensorSensitiveData censors sensitive data in complex data structures
func CensorSensitiveData(data interface{}, maskFields []string) interface{} {
	// Handle nil input
	if data == nil {
		return nil
	}

	// Use reflection to handle more dynamic type checking
	val := reflect.ValueOf(data)

	switch val.Kind() {
	case reflect.Slice:
		return censorSlice(data, maskFields)
	case reflect.Map:
		return censorMap(data, maskFields)
	case reflect.Struct:
		return censorStruct(data, maskFields)
	case reflect.Ptr:
		// Dereference pointer and recursively censor
		if val.IsNil() {
			return nil
		}
		return CensorSensitiveData(val.Elem().Interface(), maskFields)
	case reflect.String:
		return data
	default:
		return data
	}
}

// censorSlice handles censoring slice types
func censorSlice(data interface{}, maskFields []string) interface{} {
	val := reflect.ValueOf(data)
	censoredSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Len())

	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		censoredItem := CensorSensitiveData(item, maskFields)
		censoredSlice.Index(i).Set(reflect.ValueOf(censoredItem))
	}

	return censoredSlice.Interface()
}

// censorMap handles censoring map types
func censorMap(data interface{}, maskFields []string) interface{} {
	val := reflect.ValueOf(data)
	censoredMap := reflect.MakeMap(val.Type())

	iter := val.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		// Check if the key (converted to string) is in maskFields
		keyStr := fmt.Sprintf("%v", key.Interface())

		var censoredValue reflect.Value
		if contains(maskFields, keyStr) {
			// Mask the entire value if the key matches
			censoredValue = reflect.ValueOf(maskValue(value.Interface()))
		} else {
			// Recursively censor nested structures
			censoredValue = reflect.ValueOf(CensorSensitiveData(value.Interface(), maskFields))
		}

		censoredMap.SetMapIndex(key, censoredValue)
	}

	return censoredMap.Interface()
}

// censorStruct handles censoring struct types
func censorStruct(data interface{}, maskFields []string) interface{} {
	val := reflect.ValueOf(data)
	typ := val.Type()

	// Create a new struct of the same type
	censoredStruct := reflect.New(typ).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check if the field name is in maskFields
		if contains(maskFields, fieldType.Name) {
			// Mask the field value
			censoredStruct.Field(i).Set(reflect.ValueOf(maskValue(field.Interface())))
		} else {
			// Recursively censor nested fields
			censoredValue := CensorSensitiveData(field.Interface(), maskFields)
			censoredStruct.Field(i).Set(reflect.ValueOf(censoredValue))
		}
	}

	return censoredStruct.Interface()
}

// contains checks if a slice contains a given string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, item) {
			return true
		}
	}
	return false
}

// maskValue provides advanced masking for different value types
func maskValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return maskString(v)
	case fmt.Stringer:
		return maskString(v.String())
	case []byte:
		return maskString(string(v))
	case nil:
		return nil
	default:
		// For other types, use reflection to handle more cases
		return maskReflectedValue(value)
	}
}

// maskString provides sophisticated string masking
func maskString(s string) string {
	// Default masking for other strings
	if len(s) > 2 {
		maskLen := len(s) - 2
		if maskLen > 8 { // 8 because we keep first and last char
			maskLen = 8
		}
		return string(s[0]) + strings.Repeat("*", maskLen) + string(s[len(s)-1])
	}
	return strings.Repeat("*", len(s))
}

// maskReflectedValue handles masking for complex types using reflection
func maskReflectedValue(value interface{}) interface{} {
	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		// Create a masked slice of the same length
		maskedSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Len())
		for i := 0; i < val.Len(); i++ {
			maskedSlice.Index(i).Set(reflect.ValueOf("*****"))
		}
		return maskedSlice.Interface()
	case reflect.Struct:
		// Create a struct with all fields masked
		maskedStruct := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			maskedStruct.Field(i).Set(reflect.ValueOf("*****"))
		}
		return maskedStruct.Interface()
	default:
		return "*****"
	}
}

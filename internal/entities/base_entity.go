package entities

import (
	"reflect"
	"strconv"
	"time"
)

type BaseEntity struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var entityMaxLengths map[reflect.Type]map[string]int = make(map[reflect.Type]map[string]int)

func TruncateValue[T any](fieldName string, oldValue string, withDots bool) string {
	if maxLenMap, ok := entityMaxLengths[reflect.TypeFor[T]()]; ok {
		if maxLength, found := maxLenMap[fieldName]; found {
			return truncateString(oldValue, maxLength, withDots)
		}
	}
	return oldValue
}

func TruncateAllEntityFields[T any](entity *T, withDots bool) {
	t := reflect.TypeFor[T]()
	if maxLenMap, ok := entityMaxLengths[t]; ok {
		entityValue := reflect.ValueOf(entity)
		if !entityValue.IsValid() || entityValue.IsNil() {
			return
		}
		elem := entityValue.Elem()
		for fieldName, maxLength := range maxLenMap {
			field, found := t.FieldByName(fieldName)
			if !found {
				continue
			}

			fieldValue := elem.FieldByName(fieldName)
			if !fieldValue.IsValid() || !fieldValue.CanSet() {
				continue
			}

			switch field.Type.Kind() {
			case reflect.String:
				value := fieldValue.String()
				truncatedValue := truncateString(value, maxLength, withDots)
				fieldValue.SetString(truncatedValue)
			case reflect.Pointer:
				if field.Type.Elem().Kind() == reflect.String && !fieldValue.IsNil() {
					value := fieldValue.Elem().String()
					truncatedValue := truncateString(value, maxLength, withDots)
					fieldValue.Elem().SetString(truncatedValue)
				}
			}
		}
	}
}

func registerEntityMaxLengths[T any]() {
	t := reflect.TypeFor[T]()
	maxLenMap := make(map[string]int)
	for field := range t.Fields() {
		maxlenTag := field.Tag.Get("maxlen")
		if maxlenTag != "" {
			maxlen, err := strconv.Atoi(maxlenTag)
			if err == nil {
				maxLenMap[field.Name] = maxlen
			}
		}
	}
	entityMaxLengths[t] = maxLenMap
}

func truncateString(s string, maxLength int, withDots bool) string {
	runes := []rune(s)
	if len(runes) <= maxLength {
		return s
	}
	if maxLength <= 3 || !withDots {
		return string(runes[:maxLength])
	}
	return string(runes[:maxLength-3]) + "..."
}

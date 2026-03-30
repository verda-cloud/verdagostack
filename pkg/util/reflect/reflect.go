// Package reflect provides struct field inspection, selective field copy,
// and YAML-based deep copy utilities. See README.md for examples.
package reflect

import (
	"reflect"

	"gopkg.in/yaml.v3"
)

// GetFieldsMap returns a map of field names to values for the given struct.
// If fields is non-empty, only those named fields are included. Nested structs
// and pointers are expanded recursively.
func GetFieldsMap(obj any, fields []string) map[string]any {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	result := make(map[string]any, v.NumField())

	for i := range v.NumField() {
		name := t.Field(i).Name
		if len(fields) > 0 && !containsString(fields, name) {
			continue
		}

		field := v.Field(i)
		switch field.Kind() {
		case reflect.Struct, reflect.Ptr:
			result[name] = GetFieldsMap(field.Interface(), nil)
		default:
			result[name] = field.Interface()
		}
	}
	return result
}

// CopyFields copies the named fields from src to dst. Returns true if any
// field value changed. Both src and dst must be pointers to structs.
func CopyFields(src, dst any, fields []string) (changed bool, err error) {
	srcMap := GetFieldsMap(src, fields)
	dstMap := GetFieldsMap(dst, fields)
	if reflect.DeepEqual(srcMap, dstMap) {
		return false, nil
	}

	dstVal := reflect.ValueOf(dst).Elem()
	for k, v := range srcMap {
		dstVal.FieldByName(k).Set(reflect.ValueOf(v))
	}
	return true, nil
}

// CopyViaYAML marshals src to YAML and unmarshals into dst. Useful for
// copying between structs with compatible YAML tags.
func CopyViaYAML(dst, src any) error {
	if src == nil || dst == nil {
		return nil
	}
	data, err := yaml.Marshal(src)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, dst)
}

// StructName returns the type name of the given struct (or pointer to struct).
func StructName(obj any) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

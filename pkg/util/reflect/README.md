# reflect

Struct field inspection, selective field copy, and YAML-based deep copy utilities.

## Functions

| Function | Description |
|----------|-------------|
| `GetFieldsMap(obj, fields)` | Returns a map of field names → values; pass `nil` for all fields |
| `CopyFields(src, dst, fields)` | Copies named fields from src to dst; reports if anything changed |
| `CopyViaYAML(dst, src)` | Deep copy via YAML marshal/unmarshal (works across compatible types) |
| `StructName(obj)` | Returns the type name of a struct or pointer-to-struct |

## Examples

### Inspect struct fields

```go
type User struct {
    Name  string
    Email string
    Age   int
}
u := User{Name: "alice", Email: "alice@example.com", Age: 30}

all := reflect.GetFieldsMap(u, nil)
// map[Name:alice Email:alice@example.com Age:30]

subset := reflect.GetFieldsMap(u, []string{"Name", "Age"})
// map[Name:alice Age:30]
```

### Copy selected fields between structs

```go
src := &User{Name: "bob", Email: "bob@example.com", Age: 25}
dst := &User{Name: "alice", Email: "alice@example.com", Age: 30}

changed, err := reflect.CopyFields(src, dst, []string{"Name", "Age"})
// dst is now {Name:"bob", Email:"alice@example.com", Age:25}
// changed == true
```

### Deep copy via YAML serialization

Useful when copying between structs with compatible YAML tags but different Go types.

```go
type ConfigA struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}
type ConfigB struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

a := ConfigA{Host: "localhost", Port: 8080}
var b ConfigB
err := reflect.CopyViaYAML(&b, &a)
// b == ConfigB{Host:"localhost", Port:8080}
```

### Get struct type name

```go
name := reflect.StructName(&User{}) // "User"
```

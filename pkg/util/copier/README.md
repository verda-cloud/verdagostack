# copier

Deep-copy helpers with built-in type converters for `time.Time` ↔ `timestamppb.Timestamp`.

## Functions

| Function | Description |
|----------|-------------|
| `CopyWithConverters(dst, src, ...TypeConverter)` | Deep copy with time converters + optional custom converters |
| `Copy(dst, src)` | Shallow copy |
| `TypeConverters()` | Returns the built-in `time.Time` ↔ `Timestamp` converters |

## Examples

### Deep copy between DB model and API proto

```go
type DBUser struct {
    Name      string
    CreatedAt time.Time
}
type APIUser struct {
    Name      string
    CreatedAt *timestamppb.Timestamp
}

dbUser := DBUser{Name: "alice", CreatedAt: time.Now()}
var apiUser APIUser
err := copier.CopyWithConverters(&apiUser, &dbUser)
// apiUser.CreatedAt is now a *timestamppb.Timestamp
```

### Simple shallow copy

```go
var dst Config
err := copier.Copy(&dst, &src)
```

### Adding custom type converters

```go
import copierlib "github.com/jinzhu/copier"

custom := copierlib.TypeConverter{
    SrcType: MyEnum(0),
    DstType: "",
    Fn: func(src interface{}) (interface{}, error) {
        return src.(MyEnum).String(), nil
    },
}
err := copier.CopyWithConverters(&dst, &src, custom)
```

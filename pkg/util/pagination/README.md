# pagination

Helpers for page-based offset calculation.

## Functions

| Function | Description |
|----------|-------------|
| `GetPageOffset(pageNum, pageSize)` | Zero-based offset from 1-based page number and page size |

## Example

```go
// Page 1, size 20 → offset 0
offset := pagination.GetPageOffset(1, 20)

// Page 3, size 20 → offset 40
offset = pagination.GetPageOffset(3, 20)

// Use with a SQL query
db.Offset(int(pagination.GetPageOffset(req.Page, req.PageSize))).Limit(int(req.PageSize)).Find(&results)
```

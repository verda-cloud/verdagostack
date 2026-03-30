// Package pagination provides helpers for page-based offset calculation.
package pagination

// GetPageOffset returns the zero-based offset for the given 1-based page number
// and page size. For example, page 3 with size 20 returns offset 40.
func GetPageOffset(pageNum, pageSize int64) int64 {
	return (pageNum - 1) * pageSize
}

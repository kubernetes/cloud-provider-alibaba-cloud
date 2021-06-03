package alibaba

import (
	"fmt"
	"strings"
)

// A PaginationResponse represents a response with pagination information
type PaginationResult struct {
	TotalCount int
	PageNumber int
	PageSize   int
}

type Pagination struct {
	PageNumber int
	PageSize   int
}

// NextPage gets the next page of the result set
func (r *PaginationResult) NextPage() *Pagination {
	if r.PageNumber*r.PageSize >= r.TotalCount {
		return nil
	}
	return &Pagination{PageNumber: r.PageNumber + 1, PageSize: r.PageSize}
}

func formatErrorMessage(err error) error {
	if err == nil {
		return err
	}

	attrs := strings.Split(err.Error(), "\n")
	if len(attrs) != 5 {
		return err
	}
	return fmt.Errorf(strings.Join(attrs[3:], ", "))
}

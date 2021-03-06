package store

import "strings"

type Filters struct {
	Page  int    `json:"page" validate:"gt=0"`
	Limit int    `json:"limit" validate:"gt=0,lt=100"`
	Sort  string `json:"sort" validate:"oneof='id' 'title' 'payday' '-id' '-title' '-payday'"`
}

func (f Filters) sortColumn() string {
	return strings.TrimPrefix(f.Sort, "-")
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.Limit
}

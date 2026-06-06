package helper

import (
	"mysql/model"
	"mysql/request"
)

func BuildPaginationMeta(pf request.Pagination, totalCount int64) *model.PaginationMetadata {
	return &model.PaginationMetadata{
		CurrentPage: pf.Page,
		PageSize:    pf.PageSize,
		TotalCount:  totalCount,
		TotalPages:  (int(totalCount) + pf.PageSize - 1) / pf.PageSize,
	}
}

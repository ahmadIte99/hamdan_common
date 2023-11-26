package common

import (
	"strconv"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func Paginator(page string, perPage string, FindOptions *options.FindOptions) (int64, int64) {
	// page := r.URL.Query().Get("page")
	// perPage := r.URL.Query().Get("perPage")
	if perPage == "" {
		perPage = "40"
	}
	limit, _ := strconv.ParseInt(perPage, 10, 32)

	if page != "" {
		pageNum, _ := strconv.ParseInt(page, 10, 32)
		if pageNum == 1 {
			FindOptions.SetSkip(0)
			FindOptions.SetLimit(limit)
			return pageNum, limit
		}

		FindOptions.SetSkip((pageNum - 1) * limit)
		FindOptions.SetLimit(limit)
		return pageNum, limit

	}

	FindOptions.SetSkip(0)
	FindOptions.SetLimit(limit)
	return 0, limit
}

package helpers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParsePagination(c *gin.Context) (int, int) {
	limit := 100
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		if l > 500 {
			l = 500
		}
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}
	return limit, offset
}
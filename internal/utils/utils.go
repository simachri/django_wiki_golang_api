package utils

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// HandleErr returns 'true' if an error has been handled.
func HandleErr(c *gin.Context, e *error, m string) bool {
	if *e == nil {
		return false
	}
	msg := fmt.Sprintf(m, *e)
	fmt.Fprint(os.Stderr, msg)
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	return true
}


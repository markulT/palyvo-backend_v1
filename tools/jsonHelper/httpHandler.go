package jsonHelper

import "github.com/gin-gonic/gin"

var DefaultHttpErrors = map[string]ApiError{
	"BadRequest": {Err: "Bad request", Status: 400},
	"Unauthorized": {Err: "Unauthorized", Status: 401},
	"Forbidden": {Err: "Forbidden", Status: 403},
	"InternalServerError": {Err: "Internal server error", Status: 500},
}

type apiFunction func(*gin.Context) error
type ApiError struct {
	Err string `json:"error"`
	Status int `json:"status"`
}
func (e ApiError) Error() string {
	return e.Err
}

func MakeHttpHandler(f apiFunction) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err:=f(c);err!=nil {
			if e, ok := err.(*ApiError); ok {
				c.JSON(e.Status, gin.H{"error":e})
				c.Abort()
				return
			}
			c.JSON(500, gin.H{"error":err})
			c.Abort()
			return
		}
	}
}

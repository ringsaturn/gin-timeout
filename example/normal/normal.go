package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	timeout "github.com/ringsaturn/gin-timeout"
)

func main() {

	// create new gin without any middleware
	engine := gin.Default()

	defaultMsg := `{"code": -1, "msg":"http: Handler timeout"}`
	// add timeout middleware with 2 second duration
	engine.Use(timeout.Timeout(
		timeout.WithTimeout(2*time.Second),
		timeout.WithErrorHttpCode(http.StatusRequestTimeout), // optional
		timeout.WithDefaultMsg(defaultMsg),                   // optional
		timeout.WithCallBack(func(r *http.Request) {
			fmt.Println("timeout happen, url:", r.URL.String())
		}), // optional
	))
	// create a handler that will last 1 seconds
	engine.GET("/short", short)

	// create a handler that will last 5 seconds
	engine.GET("/long", long)

	// create a handler that will last 5 seconds but can be canceled.
	engine.GET("/long2", long2)

	// create a handler that will last 20 seconds but can be canceled.
	engine.GET("/long3", long3)

	engine.GET("/boundary", boundary)

	// run the server
	log.Fatal(engine.Run(":8080"))
}

func short(c *gin.Context) {
	time.Sleep(1 * time.Second)
	c.JSON(http.StatusOK, gin.H{"hello": "short"})
}

func long(c *gin.Context) {
	time.Sleep(3 * time.Second)
	c.JSON(http.StatusOK, gin.H{"hello": "long"})
}

func boundary(c *gin.Context) {
	time.Sleep(2 * time.Second)
	c.JSON(http.StatusOK, gin.H{"hello": "boundary"})
}

func long2(c *gin.Context) {
	if doSomething(c.Request.Context()) {
		c.JSON(http.StatusOK, gin.H{"hello": "long2"})
	}
}

func long3(c *gin.Context) {
	// request a slow service
	// see  https://github.com/ringsaturn/gin-timeout/blob/master/example/slow_service.go
	url := "http://localhost:8882/hello"
	// Notice:
	// Please use c.Request.Context(), the handler will be canceled where timeout event happen.
	req, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	client := http.Client{Timeout: 100 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Where timeout event happen, a error will be received.
		fmt.Println("error1:", err)
		return
	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error2:", err)
		return
	}
	fmt.Println(s)
}

// A cancelCtx can be canceled.
// When canceled, it also cancels any children that implement canceler.
func doSomething(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		fmt.Println("doSomething is canceled.")
		return false
	case <-time.After(5 * time.Second):
		fmt.Println("doSomething is done.")
		return true
	}
}

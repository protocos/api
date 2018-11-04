package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/Knetic/govaluate"
	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.POST("/math", func(c *gin.Context) {
		defer panicPrevention()

		var exp Expression
		c.Bind(&exp)
		spew.Dump(exp)

		expression, err := govaluate.NewEvaluableExpression(exp.Expression)
		if err != nil {
			spew.Dump(err)
		}
		result, err := expression.Evaluate(nil)
		if err != nil {
			spew.Dump(err)
		}

		var roundedResult string
		// if result < 1 {
		// 	roundedResult = fmt.Sprintf("%.2f", result)
		// } else {
		roundedResult = fmt.Sprintf("%."+strconv.Itoa(exp.Round)+"f", result)
		// }

		asFloat, err := strconv.ParseFloat(roundedResult, 64)
		if err != nil {
			spew.Dump(asFloat, err)
		}

		c.JSON(http.StatusOK, asFloat)
	})

	router.POST("/twitter", func(c *gin.Context) {
		defer panicPrevention()
		var twitterRequest URLRequest
		c.Bind(&twitterRequest)
		spew.Dump(twitterRequest)

		resp, err := http.Get(twitterRequest.Url)
		if err != nil {
			spew.Dump(err)
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			spew.Dump(err)
		}
		bodyString := string(bodyBytes)

		r, _ := regexp.Compile("<title>(.+?)</title>")
		matches := r.FindStringSubmatch(bodyString)
		str := matches[1]
		str = html.UnescapeString(str)

		c.JSON(http.StatusOK, str)
	})

	router.Run(":" + port)
}

type Expression struct {
	Expression string `json:"expression" binding:"required"`
	Round      int    `json:"round" binding:"required"`
}

type URLRequest struct {
	Url string `json:"url" binding:"required"`
}

type URLTagRequest struct {
	Url        string `json:"url" binding:"required"`
	Tag        string `json:"tag" binding:"required"`
	Occurrence int    `json:"occurrence" binding:"required"`
}

type Add struct {
	First  float64 `json:"first" binding:"required"`
	Second float64 `json:"second" binding:"required"`
}

func panicPrevention() {
	if err := recover(); err != nil {
		fmt.Println(err)
	}
}

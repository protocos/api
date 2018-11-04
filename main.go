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
	"strings"

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
		var request URLRequest
		c.Bind(&request)
		spew.Dump(request)

		pageSource := getWebpageSource(request)

		r, _ := regexp.Compile("<title>(.+?)</title>")
		matches := r.FindStringSubmatch(pageSource)
		str := matches[1]
		str = html.UnescapeString(str)

		c.JSON(http.StatusOK, str)
	})

	router.POST("/amazon", func(c *gin.Context) {
		defer panicPrevention()
		var request URLRequest
		c.Bind(&request)
		spew.Dump(request)

		pageSource := getWebpageSource(request)

		divRegex, _ := regexp.Compile("<div.+ebooksImageBlockContainer.+>(.|\n)*?</div>")
		matches := divRegex.FindStringSubmatch(pageSource)
		divBlock := matches[0]

		spew.Dump(divBlock)
		imgRegex, _ := regexp.Compile("<img.+src=\"((.|\n)*?)\".+>")
		matches = imgRegex.FindStringSubmatch(divBlock)
		imgBlock := matches[1]

		imgLink := html.UnescapeString(imgBlock)
		imgLink = strings.Replace(imgLink, "https://", "", -1)
		imgLink = strings.Replace(imgLink, "http://", "", -1)

		c.String(http.StatusOK, imgLink)
	})

	router.Run(":" + port)
}

func getWebpageSource(request URLRequest) string {

	resp, err := http.Get(request.Url)
	if err != nil {
		spew.Dump(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		spew.Dump(err)
	}
	bodyAsString := string(bodyBytes)

	return bodyAsString
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

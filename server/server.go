package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const categoriesURL = "https://opentdb.com/api_category.php"

//QuestionAnswers is a question and the corresponding answers
//both correct and incorrect
type QuestionAnswers struct {
	Category   string   `json:"category"`
	Correct    string   `json:"correct_answer"`
	Difficulty string   `json:"difficulty"`
	Incorrect  []string `json:"incorrect_answers"`
	Question   string   `json:"question"`
	Type       string   `json:"type"`
}

//Category represents one category of questions
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//CategoriesResponse is the structure in which we expect
//to get the categories from the trivia server
type CategoriesResponse struct {
	Categories []Category `json:"trivia_categories"`
}

//QuestionResponse represents a response for one question
type QuestionResponse struct {
	Question string `json:"question"`
	User     string `json:"user"`
	Correct  bool   `json:"correct"`
}

//Serve will start a new server for trivia
func Serve(port string) error {

	cats, err := populateCategories(categoriesURL)
	if err != nil {
		log.Fatalf("Error getting initial categories: %s", err.Error())
	}

	r := gin.Default()
	r.GET("/categories", categories(cats))
	r.Static("/trivia", "./html")
	return r.Run(port)
}

func populateCategories(url string) ([]Category, error) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	categResp := CategoriesResponse{}
	err = json.Unmarshal(body, &categResp)

	if err != nil {
		return nil, err
	}

	return categResp.Categories, nil
}

func categories(cats []Category) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, CategoriesResponse{
			Categories: cats,
		})
	}
}

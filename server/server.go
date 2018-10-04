package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

type userAnswer struct {
	User    string
	Correct bool
}

type stats struct {
	User string
	OK   int
	NOK  int
}

const categoriesURL = "https://opentdb.com/api_category.php"

var mut = sync.Mutex{}

var responses map[string][]QuestionResponse

var liveStats = map[string]stats{}

var ch = make(chan string, 10)
var done = make(chan struct{}, 1)
var chLive = make(chan userAnswer, 10)

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
	Correct  string `json:"correct"`
}

var m *melody.Melody

//Serve will start a new server for trivia
func Serve(port string) error {

	responses = map[string][]QuestionResponse{}

	cats, err := populateCategories(categoriesURL)
	if err != nil {
		log.Fatalf("Error getting initial categories: %s", err.Error())
	}

	r := gin.Default()
	m = melody.New()

	r.GET("/categories", categories(cats))
	r.POST("/report", report)
	r.GET("/summary", summary)
	r.Static("/trivia", "./html")
	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	go liveStatistics()
	go wsHandler()

	return r.Run(port)
}

func wsHandler() {
	for {
		select {
		case msg := <-ch:
			if err := m.Broadcast([]byte(msg)); err != nil {
				log.Println("Can't send")
			}
		case <-done:
			log.Println("Finished")
			break
		}
	}
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

func report(c *gin.Context) {
	r := QuestionResponse{}

	err := c.BindJSON(&r)

	if err != nil {
		log.Printf("Error reading report: %s\n", err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
	}

	mut.Lock()
	defer mut.Unlock()

	ur, ok := responses[r.User]

	if !ok {
		ur = []QuestionResponse{r}
	}

	ur = append(ur, r)

	responses[r.User] = ur

	chLive <- userAnswer{
		User:    r.User,
		Correct: r.Correct == "true",
	}

	c.JSON(200, gin.H{"status": "OK"})
}

func summary(c *gin.Context) {
	user := c.Query("user")

	if user == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	mut.Lock()
	ur, ok := responses[user]
	responses[user] = []QuestionResponse{}
	mut.Unlock()

	if !ok {
		log.Println("Unknown user: ", user)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	correct := 0
	total := 0
	for _, u := range ur {
		total = total + 1
		if u.Correct == "true" {
			correct = correct + 1
		}
	}

	c.JSON(200, gin.H{"report": fmt.Sprintf("%d/ %d", correct, total)})

}

func liveStatistics() {
	for {
		ua := <-chLive
		ls, ok := liveStats[ua.User]
		if !ok {
			ls = stats{OK: 0, NOK: 0, User: ua.User}
		}
		if ua.Correct {
			ls.OK = ls.OK + 1
		} else {
			ls.NOK = ls.NOK + 1
		}

		liveStats[ua.User] = ls

		liveUserStats := []stats{}

		for _, s := range liveStats {
			liveUserStats = append(liveUserStats, s)
		}
		sort.Slice(liveUserStats, func(i, j int) bool {
			if liveUserStats[i].OK == liveUserStats[j].OK {
				return liveUserStats[i].NOK < liveUserStats[j].NOK
			}
			return liveUserStats[i].OK > liveUserStats[j].OK
		})

		data, err := json.Marshal(liveUserStats)

		if err != nil {
			log.Printf("Cannot show stats: %s", err.Error())
			return
		}

		log.Println(string(data))
		ch <- string(data)
	}
}

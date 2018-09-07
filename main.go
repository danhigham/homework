package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	ics "github.com/PuloV/ics-golang"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type CalendarEvent struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartAt     time.Time `json:"start_at,string"`
	EndAt       time.Time `json:"end_at,string"`
	ContextCode string    `json:"context_code"`
	Hidden      bool      `json:"hidden"`
}

type Event struct {
	Start       time.Time
	End         time.Time
	Created     time.Time
	Modified    time.Time
	Description string
	Location    string
	Summary     string
	Id          string
}

func EventFromICSEvent(icsEvent *ics.Event) *Event {
	e := new(Event)
	e.Start = icsEvent.GetStart()
	e.End = icsEvent.GetEnd()
	e.Created = icsEvent.GetCreated()
	e.Modified = icsEvent.GetLastModified()
	e.Description = icsEvent.GetDescription()
	e.Location = icsEvent.GetLocation()
	e.Summary = icsEvent.GetSummary()
	e.Id = icsEvent.GetID()
	return e
}

type Course struct {
	Id       int      `json:"id"`
	Name     string   `json:"name"`
	Calendar Calendar `json:"calendar"`
}

type Calendar struct {
	Url string `json:"ics"`
}

type Assignment struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueAt       time.Time `json:"due_at,string"`
}

func main() {

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Static("./static"))

	// Routes
	e.GET("/courses.json", courses)
	e.GET("/courses/today.json", coursesToday)
	e.GET("/courses/:id/assignments.json", assignments)
	e.GET("/courses/:ids/today.json", today)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func today(c echo.Context) error {

	courseCodes := strings.Split(c.Param("ids"), ",")
	for i := range courseCodes {
		courseCodes[i] = fmt.Sprintf("context_codes[]=course_%s", courseCodes[i])
	}

	qParams := strings.Join(courseCodes, "&")
	eventsUrl := fmt.Sprintf("api/v1/calendar_events?%s", qParams)
	assignmentsUrl := fmt.Sprintf("api/v1/calendar_events?type=assignment&%s", qParams)

	events, err := getEvents(eventsUrl)

	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
	}

	assignmentEvents, err := getEvents(assignmentsUrl)

	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
	}

	for i := range assignmentEvents {
		events = append(events, assignmentEvents[i])
	}

	return c.JSON(http.StatusOK, events)
}

func getEvents(url string) ([]CalendarEvent, error) {
	var events []CalendarEvent
	resp, err := doGet(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		jsonBlob, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(jsonBlob, &events)

		if err != nil {
			return nil, err
		}
	}

	return events, nil
}

func coursesToday(c echo.Context) error {
	parser := ics.New()
	parserChan := parser.GetInputChan()
	outputChan := parser.GetOutputChan()

	var courses []Course
	var events []*Event

	resp, err := doGet("api/v1/courses")
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
	}

	if resp.StatusCode == http.StatusOK {
		jsonBlob, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(jsonBlob, &courses)

		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
		}
	}

	go func() {
		for eventIn := range outputChan {
			events = append(events, EventFromICSEvent(eventIn))
		}
	}()

	for i := range courses {
		icsUrl := courses[i].Calendar.Url
		fmt.Println(icsUrl)
		parserChan <- icsUrl
	}

	time.Sleep(1000 * time.Millisecond)

	return c.JSON(http.StatusOK, events)
}

func courses(c echo.Context) error {
	var courses []Course
	resp, err := doGet("api/v1/courses")
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
	}

	if resp.StatusCode == http.StatusOK {
		jsonBlob, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(jsonBlob, &courses)

		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
		}
	}

	return c.JSON(http.StatusOK, courses)
}

func assignments(c echo.Context) error {
	var assignments []Assignment

	resp, err := doGet(fmt.Sprintf("api/v1/courses/%s/assignments", c.Param("id")))
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
	}

	if resp.StatusCode == http.StatusOK {
		jsonBlob, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(jsonBlob, &assignments)

		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
		}
	}
	return c.JSON(http.StatusOK, assignments)
}

func doGet(url string) (*http.Response, error) {
	token := os.Getenv("CANVAS_TOKEN")
	school := os.Getenv("CANVAS_SCHOOL")

	client := &http.Client{}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s.instructure.com/%s", school, url), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return client.Do(req)
}

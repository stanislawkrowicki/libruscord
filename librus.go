package main

import (
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

const (
	baseUrl      = "https://api.librus.pl"
	primaryUrl   = "https://api.librus.pl/OAuth/Authorization?client_id=46&response_type=code&scope=mydata"
	loginUrl     = "https://api.librus.pl/OAuth/Authorization?client_id=46"
	timetableUrl = "https://synergia.librus.pl/terminarz"
)

type loginResponse struct {
	Status string `json:"status"`
	GoTo   string `json:"goTo"`
}

type Session struct {
	Client http.Client
}

type LessonEntity struct {
	Number  string
	Subject string
	Teacher string
	Group   string
	URL     string
}

func (s Session) parseLesson(node *html.Node) (LessonEntity, bool) {
	var entity LessonEntity

	numberNode, _ := htmlquery.Query(node, "//text()[1]")
	numStr := htmlquery.InnerText(numberNode)
	entity.Number = strings.ReplaceAll(strings.Split(numStr, ":")[1], " ", "")
	if entity.Number == "" { // looks like it is something else than a lesson
		return entity, false
	}

	subjectNode, _ := htmlquery.Query(node, "//span[contains(@class, 'przedmiot')]")
	entity.Subject = htmlquery.InnerText(subjectNode)

	title := htmlquery.SelectAttr(node, "title")

	r := regexp.MustCompile("Nauczyciel:(.*?)<br />")
	teacher := r.FindString(title)
	// Go Regexp does not support lookahead and lookbehind, so it has to be done explicitly
	teacher = strings.ReplaceAll(teacher, "Nauczyciel: ", "")
	teacher = strings.ReplaceAll(teacher, "<br />", "")
	entity.Teacher = teacher

	groupNode, _ := htmlquery.Query(node, "//text()[3]")
	entity.Group = htmlquery.InnerText(groupNode)

	urlNode, _ := htmlquery.Query(node, "//a")
	entity.URL = htmlquery.SelectAttr(urlNode, "href")

	return entity, true
}

func (s Session) allToday(document *html.Node) (*html.Node, error) {
	xpath := "//td[contains(@class, 'center today')]"
	list, err := htmlquery.Query(document, xpath)
	return list, err
}

func (s Session) extractLessons(dayEvents *html.Node) ([]LessonEntity, error) {
	var lessons []LessonEntity

	xpath := "/div/table/tbody/tr/td"
	events, err := htmlquery.QueryAll(dayEvents, xpath)
	if err != nil {
		return []LessonEntity{}, err
	}

	for _, event := range events {
		lesson, isLesson := s.parseLesson(event)
		if !isLesson {
			continue
		}
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func (s *Session) GetLessons() ([]LessonEntity, error) {
	timetable, err := s.Client.Get(timetableUrl)
	if err != nil {
		return nil, err
	}

	doc, err := htmlquery.Parse(timetable.Body)
	if err != nil {
		return nil, err
	}

	all, err := s.allToday(doc)
	if err != nil {
		return nil, err
	}

	if all == nil {
		return []LessonEntity{}, nil
	}

	lessons, err := s.extractLessons(all)
	if err != nil {
		return nil, err
	}

	return lessons, nil
}

func Login(login, password string) (*Session, error) {
	jar, _ := cookiejar.New(&cookiejar.Options{})

	client := http.Client{Jar: jar}

	resp, err := client.Get(primaryUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to begin authorization")
	}

	form := url.Values{}
	form.Add("action", "login")
	form.Add("login", login)
	form.Add("pass", password)

	resp, err = client.Post(loginUrl, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("client error occured while logging in")
	} else if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("invalid login or password")
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("librus responded with error while trying to log in")
	}

	var lr loginResponse
	loginResponseBody, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(loginResponseBody, &lr)
	if err != nil {
		return nil, fmt.Errorf("got invalid response from login")
	}

	grant, err := client.Get(baseUrl + lr.GoTo)
	if err != nil || grant.StatusCode != 200 {
		return nil, fmt.Errorf("failed to finish authorization")
	}

	return &Session{Client: client}, nil
}

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/PuerkitoBio/goquery"
)

type PostData struct {
	Title  string
	Author string
	Date   string
}

type Config struct {
	url               string `yaml:"url"`
	outputDestination string `yaml:"output"`
}

func main() {

	var config Config

	args := os.Args
	if len(args) == 2 {
		config.url = args[2]
		config.outputDestination = args[1]
	}

	url := os.Getenv("URL")
	output := os.Getenv("OUTPUT")

	if url != "" && output != "" {
		config.url = url
		config.outputDestination = output
	}

	if !config.isValid() {
		yamlFile, err := ioutil.ReadFile("config.yml")
		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(yamlFile, &config)
		if err != nil {
			log.Fatal(err)
		}
	}

	req, err := http.NewRequest("GET", args[2], nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	posts := []PostData{}

	doc.Find(".is_colorized").Each(func(i int, s *goquery.Selection) {
		post := PostData{
			Author: strings.TrimSpace(s.Find(".name").Text()),
			Title:  strings.TrimSpace(s.Find(".topic_title").Text()),
		}

		s.Find("span").Each(func(i int, s *goquery.Selection) {
			val, ok := s.Attr("itemprop")
			if val == "dateCreated" && ok {
				post.Date = strings.TrimSpace(s.Text())
			}
		})

		posts = append(posts, post)

	})

	switch args[1] {
	case "file":
		f, err := os.Create("test.txt")
		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		for _, post := range posts {
			_, err := fmt.Fprintln(f, fmt.Sprintf("Title: %s, Author: %s, Publish date: %s \n", post.Title, post.Author, post.Date))
			if err != nil {
				log.Fatal(err)
			}
		}
	case "console":
		for _, post := range posts {
			fmt.Println(post)
		}
	default:
		fmt.Println("choose either file or console as a first argument")
	}
}

func (c Config) isValid() bool {
	if c.url == "" || c.outputDestination == "" {
		return false
	}
	return true
}

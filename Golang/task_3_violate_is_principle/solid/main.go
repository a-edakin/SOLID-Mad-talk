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

type Config struct {
	url               string `yaml:"url"`
	outputDestination string `yaml:"output"`
}

type PostData struct {
	Title  string
	Author string
	Date   string
}

type Outputter interface {
	outputToConsole([]PostData) error
	outputToFile([]PostData) error
}

type outputter struct{}

func main() {
	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	posts, err := getPosts(config)
	if err != nil {
		log.Fatal(err)
	}

	outputter := outputter{}
	if config.outputDestination == "file" {
		err := outputter.outputToConsole(posts)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = outputter.outputToConsole(posts)
	if err != nil {
		log.Fatal(err)
	}

}

func getConfig() (Config, error) {
	args := os.Args
	if len(args) == 2 {
		return Config{args[2], args[1]}, nil
	}

	url := os.Getenv("URL")
	output := os.Getenv("OUTPUT")

	if url != "" && output != "" {
		return Config{url, output}, nil
	}

	var cnf Config

	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return cnf, err
	}

	err = yaml.Unmarshal(yamlFile, &cnf)

	return cnf, nil
}

func getPosts(config Config) ([]PostData, error) {
	req, err := http.NewRequest("GET", config.url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
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

	return posts, nil
}

func (outputter) outputToConsole(posts []PostData) error {
	for _, post := range posts {
		fmt.Println(post)
	}
	return nil
}

func (outputter) outputToFile(posts []PostData) error {
	f, err := os.Create("test.txt")
	if err != nil {
		return err
	}

	defer f.Close()

	for _, post := range posts {
		_, err := fmt.Fprintln(f, fmt.Sprintf("Title: %s, Author: %s, Publish date: %s \n", post.Title, post.Author, post.Date))
		if err != nil {
			return err
		}
	}

	return nil
}

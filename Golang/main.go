package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/PuerkitoBio/goquery"
)

type PostData struct {
	Title  string
	Author string
	Date   string
}

func main() {
	args := os.Args

	if len(args) < 2 {
		log.Fatal("need to provide a proper URL to parse")
	}

	req, err := http.NewRequest("GET", args[1], nil)
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

	for _, post := range posts {
		fmt.Println(post)
	}
}

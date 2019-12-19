package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
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

func (post PostData) String() string {
	return fmt.Sprintf("Title: %s, Author: %s, Publish date: %s \n", post.Title, post.Author, post.Date)
}

type Config struct {
	url               string `yaml:"url"`
	outputDestination string `yaml:"output"`
	telegramChatID    int64  `yaml:"chat_id"`
	telegramBotToken  string `yaml:"bot_token"`
}

func main() {

	cnf, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("GET", cnf.url, nil)
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

	switch cnf.outputDestination {
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
	case "telegram":
		var msg string
		for _, post := range posts {
			msg += post.String()
		}

		var jsonStr = []byte(fmt.Sprintf(`{"chat_id": "%d", "text": "%s", "disable_notification": true}`, cnf.telegramChatID, msg))
		url := fmt.Sprintf("https://api.telegram.org/bot$%s/sendMessage", cnf.telegramBotToken)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			log.Fatal(fmt.Errorf("failed to send telegram message, status code %d", resp.StatusCode))
		}
	default:
		fmt.Println("choose either file or console as a first argument")
	}
}

func getConfig() (Config, error) {
	args := os.Args
	if len(args) == 4 {
		chatID, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return Config{}, err
		}
		return Config{args[2], args[1], chatID, args[4]}, nil
	}

	url := os.Getenv("URL")
	output := os.Getenv("OUTPUT")
	chatID := os.Getenv("CHAT_ID")
	botToken := os.Getenv("BOT_TOKEN")

	if url != "" && output != "" && chatID != "" && botToken != "" {
		id, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
		if err != nil {
			return Config{}, err
		}
		return Config{url, output, id, botToken}, nil
	}

	var cnf Config

	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return cnf, err
	}

	err = yaml.Unmarshal(yamlFile, &cnf)

	return cnf, nil
}

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

type Config struct {
	url               string `yaml:"url"`
	outputDestination string `yaml:"output"`
	telegramChatID    int64  `yaml:"chat_id"`
	telegramBotToken  string `yaml:"bot_token"`
}

type PostData struct {
	Title  string
	Author string
	Date   string
}

func (post PostData) String() string {
	return fmt.Sprintf("Title: %s, Author: %s, Publish date: %s \n", post.Title, post.Author, post.Date)
}

type Outputter interface {
	output([]PostData) error
}

type TelegramClient interface {
	send(int64, string) error
}

type telegramClient struct {
	botToken string
}

type ConsoleOutputter struct{}
type FileOutputter struct{}
type TelegramOutputter struct {
	client TelegramClient
	chat   int64
}

func main() {
	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	var outputters = map[string]Outputter{
		"console": &ConsoleOutputter{},
		"file":    &FileOutputter{},
		"telegram": &TelegramOutputter{
			chat: config.telegramChatID,
			client: telegramClient{
				botToken: config.telegramBotToken,
			},
		},
	}

	posts, err := getPosts(config)
	if err != nil {
		log.Fatal(err)
	}

	outputer, ok := outputters[config.outputDestination]
	if !ok {
		log.Fatal("need to select proper diplay type: file | console | telegram")
	}

	if err = outputer.output(posts); err != nil {
		log.Fatal(err)
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

func (*ConsoleOutputter) output(posts []PostData) error {
	for _, post := range posts {
		fmt.Println(post)
	}
	return nil
}

func (*FileOutputter) output(posts []PostData) error {
	f, err := os.Create("test.txt")
	if err != nil {
		return err
	}

	defer f.Close()

	for _, post := range posts {
		_, err := fmt.Fprintln(f, post.String())
		if err != nil {
			return err
		}
	}

	return nil
}

func (to *TelegramOutputter) output(posts []PostData) error {
	var msg string
	for _, post := range posts {
		msg += post.String()
	}
	return to.client.send(to.chat, msg)
}

func (tc telegramClient) send(chatID int64, message string) error {
	var jsonStr = []byte(fmt.Sprintf(`{"chat_id": "%d", "text": "%s", "disable_notification": true}`, chatID, message))
	url := fmt.Sprintf("https://api.telegram.org/bot$%s/sendMessage", tc.botToken)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to send telegram message, status code %d", resp.StatusCode)
	}
	return nil
}

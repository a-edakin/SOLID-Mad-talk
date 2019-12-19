# !/usr/bin/env python

import json
import os
import sys

import requests
import telebot
from bs4 import BeautifulSoup


def args_loader():
    if len(sys.argv) == 3:
        return Config(sys.argv[1], sys.argv[2], sys.argv[2:5])


def env_loader():
    output_type = os.environ.get('OUTPUT_TYPE')
    posts_url = os.environ.get('POSTS_URL')
    telegram_token = os.environ.get('TELEGRAM_TOKEN')
    chat_id = os.environ.get('CHAT_ID')

    if output_type and posts_url:
        return Config(output_type, posts_url, telegram_token, chat_id)


def json_loader():
    with open('config.json') as json_string:
        raw_config = json.loads(json_string.read())
        if raw_config.get('output_type') and raw_config.get('posts_url'):
            return Config(raw_config.pop('output_type'), raw_config.pop('posts_url'), **raw_config)


loaders = [
    args_loader,
    env_loader,
    json_loader,
]


class Config:
    def __init__(self, output_type: str, posts_url: str, telegram_token=None, chat_id=None):
        self.output_type = output_type
        self.posts_url = posts_url
        self.telegram_token = telegram_token
        self.chat_id = chat_id


def load_config() -> Config:
    for loader in loaders:
        config = loader()
        if config:
            return config


config = load_config()


def get_posts(url):
    response = requests.get(url)
    soup = BeautifulSoup(response.content, 'html.parser')

    posts = []
    for row in soup.find_all('tr', {'itemtype': 'http://schema.org/Article'}):
        posts.append({
            'title': row.find('a', {'itemprop': 'url'}).attrs.get('title'),
            'author': row.find('span', {'itemprop': 'name'}).string,
            'date': row.find('span', {'itemprop': 'dateCreated'}).string,
        })

    return posts


class TelegramBaseClient:
    def __init__(self, token: str, chat_id: str):
        self.token = token
        self.chat_id = chat_id

    def send_message(self, posts: list):
        raise NotImplemented()


class TelegramClient(TelegramBaseClient):

    def get_bot(self):
        return telebot.TeleBot(self.token)

    def send_message(self, posts: list):
        bot = self.get_bot()
        message = ''
        for post in posts:
            message += f'Title: {post["title"]} \n'
            message += f'Author {post["author"]}, Date {post["date"]}\n\n'
        bot.send_message(self.chat_id, message)


class Output:
    def __init__(self, posts):
        self.posts = posts

    def output(self):
        raise NotImplemented()


class DisplayPosts(Output):
    def output(self):
        for post in self.posts:
            print('Title: ', post['title'])
            print(f'Author {post["author"]}, Date {post["date"]}')
            print()


class SavePosts(Output):
    def output(self):
        with open('Posts.txt', 'w') as file:
            for post in self.posts:
                file.write(f'Title: {post["title"]} \n')
                file.write(f'Author {post["author"]}, Date {post["date"]}\n\n')


class SendToChat(Output):
    def output(self):
        client = TelegramClient(config.telegram_token, config.chat_id)
        client.send_message(self.posts)


output_types = {
    'console': DisplayPosts,
    'file': SavePosts,
    'telegram': SendToChat,
}

if __name__ == '__main__':
    posts_data = get_posts(config.posts_url)

    output_handler = output_types.get(config.output_type)

    output_handler(posts_data).output()

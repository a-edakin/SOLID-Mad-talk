# !/usr/bin/env python

import json
import os
import sys

import requests
from bs4 import BeautifulSoup


def args_loader():
    if len(sys.argv) == 3:
        return Config(sys.argv[1], sys.argv[2])


def env_loader():
    output_type = os.environ.get('OUTPUT_TYPE')
    posts_url = os.environ.get('POSTS_URL')

    if output_type and posts_url:
        return Config(output_type, posts_url)


def yaml_loader():
    with open('config.json') as json_string:
        raw_config = json.loads(json_string.read())
        if raw_config['output_type'] and raw_config['posts_url']:
            return Config(raw_config['output_type'], raw_config['posts_url'])


loaders = [
    args_loader,
    env_loader,
    yaml_loader,
]


class Config:
    def __init__(self, output_type: str, posts_url: str):
        self.output_type = output_type
        self.posts_url = posts_url


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


output_types = {
    'console': DisplayPosts,
    'file': SavePosts,
}

if __name__ == '__main__':
    posts_data = get_posts(config.posts_url)

    output_handler = output_types.get(config.output_type)

    output_handler(posts_data).output()

# !/usr/bin/env python

import sys

import requests
from bs4 import BeautifulSoup


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


def display_posts(posts):
    for post in posts:
        print('Title: ', post['title'])
        print(f'Author {post["author"]}, Date {post["date"]}')
        print()


if __name__ == '__main__':
    posts_url = sys.argv[1]

    posts_data = get_posts(posts_url)

    display_posts(posts_data)


#!/usr/bin/env python

import sys

import requests
from bs4 import BeautifulSoup

response = requests.get(sys.argv[2])

soup = BeautifulSoup(response.content, 'html.parser')

posts = []
for row in soup.find_all('tr', {'itemtype': 'http://schema.org/Article'}):
    posts.append({
        'title': row.find('a', {'itemprop': 'url'}).attrs.get('title'),
        'author': row.find('span', {'itemprop': 'name'}).string,
        'date': row.find('span', {'itemprop': 'dateCreated'}).string,
    })

if sys.argv[1] == 'console':
    for post in posts:
        print('Title: ', post['title'])
        print(f'Author {post["author"]}, Date {post["date"]}')
        print()

elif sys.argv[1] == 'file':
    with open('Posts.txt', 'w') as file:
        for post in posts:
            file.write(f'Title: {post["title"]} \n')
            file.write(f'Author {post["author"]}, Date {post["date"]}\n\n')


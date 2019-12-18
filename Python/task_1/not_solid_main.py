#!/usr/bin/env python

import sys

import requests
from bs4 import BeautifulSoup

response = requests.get(sys.argv[1])

soup = BeautifulSoup(response.content, 'html.parser')

for row in soup.find_all('tr', {'itemtype': 'http://schema.org/Article'}):
    author = row.find('span', {'itemprop': 'name'}).string
    date_created = row.find('span', {'itemprop': 'dateCreated'}).string
    title = row.find('a', {'itemprop': 'url'}).attrs.get('title')
    print('Title ', title)
    print(f'Author {author}, Date {date_created}')
    print()
#!/usr/bin/env python
import json
import os
import sys
import telebot

import requests
from bs4 import BeautifulSoup

output_type = ''
posts_url = ''
telegram_token = ''
chat_id = ''

if len(sys.argv) == 3:
    output_type = sys.argv[1]
    posts_url = sys.argv[2]
elif os.environ.get('OUTPUT_TYPE') and os.environ.get('POSTS_URL'):
    output_type = os.environ.get('OUTPUT_TYPE')
    posts_url = os.environ.get('POSTS_URL')
else:
    with open('config.json') as json_string:
        raw_config = json.loads(json_string.read())
        if raw_config['output_type'] and raw_config['posts_url']:
            output_type = raw_config['output_type']
            posts_url = raw_config['posts_url']

response = requests.get(posts_url)

soup = BeautifulSoup(response.content, 'html.parser')

posts = []
for row in soup.find_all('tr', {'itemtype': 'http://schema.org/Article'}):
    posts.append({
        'title': row.find('a', {'itemprop': 'url'}).attrs.get('title'),
        'author': row.find('span', {'itemprop': 'name'}).string,
        'date': row.find('span', {'itemprop': 'dateCreated'}).string,
    })

if output_type == 'console':
    for post in posts:
        print('Title: ', post['title'])
        print(f'Author {post["author"]}, Date {post["date"]}')
        print()

elif output_type == 'file':
    with open('Posts.txt', 'w') as file:
        for post in posts:
            file.write(f'Title: {post["title"]} \n')
            file.write(f'Author {post["author"]}, Date {post["date"]}\n\n')

elif output_type == 'telegram':
    bot = telebot.TeleBot(telegram_token)
    message = ''
    for post in posts:
        message += f'Title: {post["title"]} \n'
        message += f'Author {post["author"]}, Date {post["date"]}\n\n'
    bot.send_message(chat_id, message)


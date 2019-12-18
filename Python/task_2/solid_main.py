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
    output_type = sys.argv[1]
    posts_url = sys.argv[2]

    posts_data = get_posts(posts_url)

    output_handler = output_types.get(output_type)

    output_handler(posts_data).output()

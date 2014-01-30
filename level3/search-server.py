#!/usr/bin/env pypy

# This code is horrible, with if statements and global state all around
# and it's elegible for the first prize of the ugliest code I've ever
# written.

import re
import os
import sys
import json
import requests
from collections import defaultdict
from bottle import route, run, request

RE_EXTRACT_WORDS = re.compile(r"[\w]+")

INDEXED = False
NUMNODES = 3
FILELIST = []
IS_MASTER = False
SERVER_ID = None
PATH = None
BASEPORT = 9090
INDEX = defaultdict(set)
ALIASES = defaultdict(set)

WORDMAP = {}
FILEMAP = {}
FILEMAP_REV = {}

HTTP_CONNS = []


def get_word_id(word):
    global WORDMAP
    try:
        return WORDMAP[word]
    except KeyError:
        v = len(WORDMAP) + 1
        WORDMAP[word] = v
        return v


def get_file_id(path):
    global FILEMAP
    global FILEMAP_REV
    try:
        return FILEMAP[path]
    except KeyError:
        v = len(FILEMAP) + 1
        FILEMAP[path] = v
        FILEMAP_REV[v] = path
        return v


def get_file_from_id(file_id):
    return FILEMAP_REV[file_id]


def getfiles(path):
    return [
        os.path.join(root_, name)[len(path) + 1:]
        for root_, dirs, files in os.walk(path)
        for name in files
    ]


def find_word(word):
    """
    :type word: str
    """
    global INDEX
    global ALIASES
    res = set()
    for w in [w for w in ALIASES[get_word_id(word)]]:
        for file_id, line_no in INDEX[w]:
            res.add("%s:%s" % (get_file_from_id(file_id), line_no))
    return res


def load_preprocessed():
    global ALIASES
    global WORDMAP

    for line in file("wordmap.txt"):
        word, wid = line.split()
        WORDMAP[word] = int(wid)

    for line in file("preprocessed.txt"):
        word_id, is_contained_in = line.split(":")
        word_id = int(word_id)
        # include the word itself
        is_contained_in = is_contained_in.split() + [word_id]
        ALIASES[word_id] = set([int(w) for w in is_contained_in])


def create_index():
    global FILELIST
    global INDEX
    global ALIASES
    load_preprocessed()
    for path in FILELIST:
        file_id = get_file_id(path)
        full_path = os.path.join(PATH, path)

        for idx, line in enumerate(file(full_path)):
            for word in RE_EXTRACT_WORDS.findall(line):
                if len(word) > 3:
                    INDEX[get_word_id(word)].add((file_id, idx + 1))


@route("/healthcheck")
def healtcheck():
    return {"success": True}


@route("/index")
def index():
    global INDEXED
    global FILELIST
    global PATH

    path = request.query.path
    PATH = path

    if IS_MASTER:
        for i in range(NUMNODES):
            url = "http://localhost:%d/index?path=%s" % (BASEPORT + i + 1, path)
            HTTP_CONNS[i].get(url)
    else:
        FILELIST = partition_files(SERVER_ID, getfiles(path))
        create_index()

    INDEXED = True
    return {"success": True}


@route("/isIndexed")
def is_indexed():
    global INDEXED
    return {"success": INDEXED}


@route("/")
def root():
    query = request.query.q
    results = set()

    if IS_MASTER:
        for i in range(NUMNODES):
            conn = HTTP_CONNS[i]

            url = "http://localhost:%d/?q=%s" % (BASEPORT + i + 1, query)
            res = json.loads(conn.get(url).text)
            for item in res["results"]:
                results.add(item)
    else:
        results = find_word(query)

    val = {
        "success": True,
        "results": list(results)
    }
    #print val
    return val


def partition_files(node_id, filenames):
    """Return all the files that should be indexed by this node"""
    return [f for f in filenames if partition_key(f) == (node_id - 1)]


def partition_key(string):
    return hash(string) % NUMNODES


if __name__ == "__main__":
    if "--master" in sys.argv:
        IS_MASTER = True
        HTTP_CONNS = [
            requests.Session(),
            requests.Session(),
            requests.Session(),
        ]
    else:
        id_idx = sys.argv.index("--id")
        SERVER_ID = int(sys.argv[id_idx + 1])
        BASEPORT += SERVER_ID

    print ("master=%s id=%s port=%s" % (IS_MASTER, SERVER_ID, BASEPORT))
    run(port=BASEPORT)

#!/usr/bin/env pypy

# this version gave me ~100k hash/s with pypy

import hashlib
import sys
import os
import threading
import time
import uuid

MESSAGE = """tree %(tree)s
parent %(parent)s
author krat <me@example.com> %(timestamp)s +0100
committer krat <me@example.com> %(timestamp)s +0100

Give me a Gitcoin

%(threadseed)s
%(counter)s"""


def makehash(tree, parent, timestamp, threadseed, ctr):
    data = MESSAGE % {
        "tree": tree,
        "parent": parent,
        "timestamp": timestamp,
        "threadseed": threadseed,
        "counter": ctr,
    }
    s = hashlib.sha1()
    s.update("commit %u\0" % len(data))
    s.update(data)
    return s.hexdigest(), data


found = threading.Event()
found.clear()
found_h = None
found_msg = None
iter_count = 0


class Hasher(threading.Thread):
    def __init__(self, tree, parent, timestamp, difficulty, *args, **kwargs):
        threading.Thread.__init__(self, *args, **kwargs)
        self.tree = tree
        self.parent = parent
        self.timestamp = timestamp
        self.difficulty = difficulty
        self.unique = str(uuid.uuid4())

    def run(self):
        global iter_count
        global found_h
        global found_msg

        i = 0
        while not found.is_set():
            i += 1
            iter_count += 1
            h, msg = makehash(self.tree, self.parent, self.timestamp, i, self.unique)
            if h < self.difficulty:
                found.set()
                found_h = h
                found_msg = msg

if __name__ == "__main__":
    cwd = sys.argv[1]
    tree = sys.argv[2]
    parent = sys.argv[3]
    timestamp = sys.argv[4]
    difficulty = sys.argv[5]
    os.chdir(cwd)

    threads = [
        Hasher(tree, parent, timestamp, difficulty),
        Hasher(tree, parent, timestamp, difficulty),
        Hasher(tree, parent, timestamp, difficulty),
        #Hasher(tree, parent, timestamp, difficulty),
        #Hasher(tree, parent, timestamp, difficulty),
        #Hasher(tree, parent, timestamp, difficulty),
        #Hasher(tree, parent, timestamp, difficulty),
    ]

    sys.stderr.write("launching %d threads\n" % len(threads))
    for thread in threads:
        thread.start()

    start = time.time()
    while not found.is_set():
        elapsed = int(time.time() - start)
        if elapsed > 120:
            sys.stderr.write("time out, preparing to leave\n")
            found.set()

        for thread in threads:
            thread.join(3)

        elapsed = int(time.time() - start)
        hash_s = iter_count / elapsed if elapsed > 0 else 0
        sys.stderr.write("iter: %d (%d hash/s)\n" % (iter_count, hash_s))

    if found_h:
        f = open("msg", "w")
        f.write(found_msg)
        f.close()

        f = open("hash", "w")
        f.write(found_h)
        f.close()

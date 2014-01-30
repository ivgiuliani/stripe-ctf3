#!/usr/bin/env python

from collections import defaultdict

WORDMAP = {}


def get_word_id(word):
    global WORDMAP
    try:
        return WORDMAP[word]
    except KeyError:
        v = len(WORDMAP) + 1
        WORDMAP[word] = v
        return v


words = [line.strip() for line in file("dictionary") if len(line.strip()) > 3]
lengths = defaultdict(list)
for word in words:
    lengths[len(word)].append(word)


f = open("wordmap.txt", "w")
for word in words:
    wid = get_word_id(word)
    f.write(word + " " + str(wid) + "\n")
f.close()


f = open("preprocessed.txt", "w")
lengths_keys = lengths.keys()
for word in words:
    available_lengths = [length for length in lengths_keys if length > len(word)]
    is_contained_in = (str(get_word_id(w)) for l in available_lengths for w in lengths[l] if word in w)
    f.write("%d:%s\n" % (get_word_id(word), " ".join(is_contained_in)))
f.close()

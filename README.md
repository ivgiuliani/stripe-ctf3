My solutions to stripe's CTF3
=============================================================================

This repository contains my solutions for stripe's CTF 3.


Level 0
-----------------------------------------------------------------------------
The first problem was to write a program that would wrap every lowercase
version of words (passed on the stdin as a text block) not present in a
given dictionary between '<' and '>'.

I initially solved this with a one-line change to the initial ruby solution;
that didn't score very well but was enough to (barely) pass the level.
Ruby is out of my comfort zone, so I wrote another solution in Python which
is my favorite language and managed to squeeze some optimizations in (like
preprocessing the dictionary, which as it turns out --and stated in the
README-- it was always the same). This got me about 200 points and it stayed
that way until I felt comfortable enough to rewrite the whole thing in Go.
I used Go before for small projects but I didn't really know the language.
However, after I spent some time on level4, I came back to level0, rewrote
the solution in Go and managed to get to 592 points which really says a lot
about how powerful Go as a language is. The solution included here is my
final solution in Go.

It turns out that folks have done awesome things with this level to save
even a few microseconds like [injecting preprocessed hash tables in the
binary after the compilation](muehe.org/posts/stripe-ctf-3-writeup/) or
or bloom filters.


Level 1
-----------------------------------------------------------------------------
This one was interesting. I know git, but not that well, and this level asked
to generate git commits whose sh1 was smaller than a given value. They
provided you with a bash script that looped over `git hash-object` but
of course that was very slow. I wrote my initial solution in Python and
passed the level after a couple of minutes of mining doing ~100k hash/sec
on an Intel i5.

This level also included a PvP mode where players would compete against each
other. So I rewrote the solution in Java and managed to get up to 400k
hash/sec. I launched the miner on a crappy server I own during the night that
made ~50k hash/second and woke up to 880 points in the morning (I'm still not
sure how that happened). Anyway after that people started throwing hardware
at the problem (GPU miners, specifically) and I was out of the game at that
point. All I have are cheap Intel GPUs and I'm not really interested in that
kind of thing.


Level 2
-----------------------------------------------------------------------------
This was the level I liked the least, but probably for the wrong reasons.
You had to write a *shield* against a DDoS. I just counted the number of
requests and if the same ip sent more than 5 requests I'd block it.
This was enough to pass the level but it certainly wasn't the best of the
solutions.
I wanted to spend some more time on it, but the starting solution was written
in NodeJS and I despise everything written in Javascript. I hate Javascript,
it's a horrible language and the event driven paradigm that NodeJS empowers,
while it certainly is powerful, is also very hard to follow and to debug.


Level 3
-----------------------------------------------------------------------------
This one was interesting. You had to write a simple search engine for full
words (no substrings) in a random text corpora. Their starting solution was
in Scala and I spent almost two days trying to make it work and in the end
I managed to have a working solution that was enough to pass the level.
As I wasn't very familiar with Scala I rewrote the whole thing in Python.
Now, the outcome isn't certainly something I'm proud of, it's probably one
of the ugliest programs I've ever written but I hacked the solution in a
couple of hours and I jumped from ~500 points to 2065 points (using PyPy).


Level 4
-----------------------------------------------------------------------------
Good thing I didn't spend much time on the previous levels. This one asked to
write a load-balanced SQL database that supported distributed writes on a
lossy network (simulated using [octopus](https://github.com/stripe-ctf/octopus)).
In order to do so I had to explore uncharted territories, like consensus
algorithms and Go. I think I spent 3 days on it before actually arriving to
a working solution and spent the remaining 2 days on optimizing it, finally
getting to the 13th place in the leaderboard with 5649 points.

I wrote some toy programs in Go before, but this was my first attempt at
using it for anything serious. I have to say that I really like the language
and I'll try to use it more often from now on. I liked it so much that I
even went back to level0 and solved it with Go.

One minor problem I had with it was that my local solution performed quite
poorly while it performed 10 times better remotely. The last version of the
solution I wrote would get at most 100-120 correct queries locally, while
I consistenly got >1000 correct queries on stripe's servers. This was in
contrast with what everybody else was seeing (people had the opposite
problem, where remote solution was slow and local much faster) and I'm still
not sure what was causing it. Initially I thought it was because of sqlite3
(which was used as a backend) that maybe was holding me back and/or related
disk accessed, but then I ditched sqlite altogether and still experienced the
problem. At this point I can only assume it's some issue with unix sockets
that octopus was using to simulate the lossy network.

This level was the level I liked the most. It opened a lot of possible
optimizations in many areas since the score was a combination of how many
bytes you sent/received to/from other nodes and the number of queries
executed, hence to get a better score you could either optimize the exchange
protocol or improve the actual program's performance (or both).


Conclusions
-----------------------------------------------------------------------------
I like the contest and can't really complain on anything. Just a few itches:
- some folks complained that $a_given_level was not written in the
  $their_favorite_language so they rewrote the whole thing in
  $their_favorite_language. I don't think that should be the spirit of the
  CTF. For this I written Ruby, Python, Javascript, Go, Scala, Java and even
  attempted a C solution for level0, and I'm really only familiar with Python
  and Java among these (and C, but haven't touched the language in decades).
  Now it might be by chance, but I noticed a pattern, where NodeJS folks
  always rewrote everything in NodeJS; I believe that's just plain wrong.
- other folks spent more time trying to hack around the system than
  actually solving the problems. If they had put as much effort into trying
  to solve the actual problem as they did in circumventing it, they would
  be better programmers by now.
  One guy (that I know of) also downloaded each reference test case and
  essentially wrote a O(1) solution that just returned the expected output
  for level3. I'm not sure how to react to this, but it just seems unfair.
  I mean, you have an awesome score, but you didn't solve the level's
  problem. I wouldn't feel good about it.

In the end I ended up 13th, which is a great result (for me, at least).
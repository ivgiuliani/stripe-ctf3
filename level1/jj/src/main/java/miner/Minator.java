package miner;

import com.google.common.base.Charsets;
import com.google.common.collect.Lists;
import com.google.common.hash.HashCode;
import com.google.common.hash.HashFunction;
import com.google.common.hash.Hashing;

import java.io.FileNotFoundException;
import java.io.PrintWriter;
import java.io.UnsupportedEncodingException;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicLong;

public class Minator {
  private static final HashFunction HASH_F = Hashing.sha1();
  private AtomicBoolean found = new AtomicBoolean(false);
  private AtomicLong iterations = new AtomicLong(0);

  private volatile String foundMessage = null, foundHash = null;

  private final String tree;
  private final String parent;
  private final String timestamp;
  private final String difficulty;

  private static final String MESSAGE =
      "tree %s\n" +
      "parent %s\n" +
      "author krat <me@example.com> %s +0100\n" +
      "committer krat <me@example.com> %s +0100\n\n" +
      "Give me a Gitcoin\n\n" +
      "%s\n" +
      "%d";

  public final class Miner implements Runnable {
    public long counter = 0;
    private final String THREAD_ID = UUID.randomUUID().toString();

    @Override
    public void run() {
      String message, hash;

      while (!found.get()) {
        iterations.incrementAndGet();
        counter++;

        message = buildMessage(THREAD_ID, counter);
        hash = hash(message);

        if (hash.compareTo(difficulty) <= 0) {
          foundMessage = message;
          foundHash = hash;

          found.set(true);
          break;
        }
      }
    }
  }

  public static void main(String[] args) throws FileNotFoundException, UnsupportedEncodingException {
    Minator jj = new Minator(
        args[0], // tree
        args[1], // parent
        args[2], // timestamp
        args[3] // difficulty
    );
    jj.start();
  }

  Minator(String tree, String parent, String timestamp, String difficulty) {
    this.tree = tree;
    this.parent = parent;
    this.timestamp = timestamp;
    this.difficulty = difficulty;
  }

  public void start() throws FileNotFoundException, UnsupportedEncodingException {
    List<Thread> threads = Lists.newArrayList();

    for (int i = 0; i < 4; i++) {
      Thread t = new Thread(new Miner());
      threads.add(t);
      t.start();
    }

    try {
      Thread.sleep(1000);
    } catch (InterruptedException e) {
      e.printStackTrace();
    }

    long start = System.currentTimeMillis();
    long elapsed = 0;

    while (!found.get()) {
      if ((elapsed / 1000) > 45) {
        System.err.println("timeout.");
        found.set(true);
        break;
      }

      for (Thread t : threads) {
        try {
          if (t.isAlive()) {
            t.join(1500);
          } else {
            System.err.println("WARN: thread not alive");
          }
        } catch (InterruptedException e) {
          e.printStackTrace();
        }
      }

      elapsed = System.currentTimeMillis() - start;
      System.err.println(String.format("%d hash/s (elapsed: %d seconds)",
          (int)(iterations.get() / (elapsed / 1000.0f)),
          (elapsed / 1000)));
    }

    PrintWriter writer;
    if (foundHash != null && foundMessage != null) {
      writer = new PrintWriter("hash", Charsets.US_ASCII.displayName());
      writer.print(foundHash);
      writer.close();

      writer = new PrintWriter("msg", Charsets.US_ASCII.displayName());
      writer.print(foundMessage);
      writer.close();
    }
  }

  String buildMessage(String threadId, long ctr) {
    return String.format(MESSAGE, tree, parent, timestamp, timestamp, threadId, ctr);
  }

  static String hash(String msg) {
    String header = String.format("commit %d\0", msg.length());
    HashCode hc = HASH_F.newHasher()
        .putString(header + msg, Charsets.US_ASCII)
        .hash();

    return hc.toString();
  }
}

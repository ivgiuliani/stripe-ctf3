package miner;

import junit.framework.TestCase;

public class TestHash extends TestCase {
  public void testHash() {
    final String expectedHash = "735f174af3252348ac42b611844bfb611a9e3dd3";
    final String expectedMsg = "tree c999d14682a0c770c066f4ece233a16de1b7afdb\n" +
                               "parent 0000000bb75b38280c18a0f21a069d80d7e5923c\n" +
                               "author krat <me@example.com> 1390568196 +0100\n" +
                               "committer krat <me@example.com> 1390568196 +0100\n\n" +
                               "Give me a Gitcoin\n\n" +
                               "cf3560af-c03f-4963-8251-cd5f5144d14c\n" +
                               "2802978";

    Minator m = new Minator(
        "c999d14682a0c770c066f4ece233a16de1b7afdb",
        "0000000bb75b38280c18a0f21a069d80d7e5923c",
        "1390568196",
        "0000001fffffffffffffffffffffffffffffffff"
    );

    String msg = m.buildMessage("cf3560af-c03f-4963-8251-cd5f5144d14c", 2802978);
    assertEquals(expectedMsg, msg);
    assertEquals(expectedHash, Minator.hash(msg));
  }
}

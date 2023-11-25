## Conversation

This example shows a threaded conversation with elements being multi-line.

```sh
 $ go run . -depth 3
╻
┃ Root node
┃ Sphinx of black quartz, judge my vow!
┃ The quick brown fox jumps over the lazy dog.
╹
┃ ╻
┃ ┃ Child node
┃ ┃ Sphinx of black quartz, judge my vow!
┃ ┃ The quick brown fox jumps over the lazy dog.
┃ ╹
┃ ┃ ╻
┃ ┃ ┃ Child node
┃ ┃ ┃ Sphinx of black quartz, judge my vow!
┃ ┃ ┃ The quick brown fox jumps over the lazy dog.
┃ ┃ ╹
```

Or with colours:

![Screenshot](./screenshot.png)

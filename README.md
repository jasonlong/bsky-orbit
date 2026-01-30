# bsky-orbit 🔭

Discover who to follow on Bluesky based on your network.

Analyzes your follows' follows to find accounts you're not following but probably should be, ranked by how many of your follows also follow them.

## Install

Download the latest binary from [Releases](https://github.com/jasonlong/bsky-orbit/releases), or build from source:

```bash
go install github.com/jasonlong/bsky-orbit@latest
```

## Usage

```bash
bsky-orbit <handle>
```

Examples:
```bash
bsky-orbit jasonlong.me
bsky-orbit @username.bsky.social
```

No API key required.

## How it works

1. Fetches everyone you follow
2. For each person you follow, fetches who *they* follow
3. Counts how many times each account appears
4. Filters out accounts you already follow
5. Returns the top 30, ranked by "in common" count

The idea: if 50 of your follows all follow the same person, you probably want to follow them too.

## Output

The tool outputs:
- A formatted table in the terminal
- `bsky-orbit-{handle}.json` - Full data as JSON
- `bsky-orbit-{handle}.md` - Markdown table

Example:

```
🌟 TOP RECOMMENDATIONS FOR @jasonlong.me
============================================================

Rank  Account                      Followers    In Common
------------------------------------------------------------
1     Jay 🦋                       595.3K       55
2     Brian Lovin                  6.6K         50
3     dan                          62.2K        47
4     Alexandria Ocasio-Cortez     2.2M         45
5     Sarah Drasner                47.7K        45
...
------------------------------------------------------------

'In Common' = how many of your follows also follow this account
```

Markdown output (`bsky-orbit-jasonlong.me.md`):

| Rank | Account | Followers | In Common | Bio |
|------|---------|-----------|-----------|-----|
| 1 | [Jay 🦋](https://bsky.app/profile/jay.bsky.team) | 595.3K | 55 | CEO of Bluesky, steward of AT Protocol. |
| 2 | [Brian Lovin](https://bsky.app/profile/brianlovin.com) | 6.6K | 50 | Co-founder of Campsite |
| 3 | [dan](https://bsky.app/profile/danabra.mov) | 62.2K | 47 | falling down in the green grass |
| ... | | | | |

## Building from source

```bash
git clone https://github.com/jasonlong/bsky-orbit
cd bsky-orbit
go build -o bsky-orbit .
```

### Cross-compile for all platforms

```bash
# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bsky-orbit-darwin-arm64 .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o bsky-orbit-darwin-amd64 .

# Linux
GOOS=linux GOARCH=amd64 go build -o bsky-orbit-linux-amd64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o bsky-orbit-windows-amd64.exe .
```

## Why no API key?

Bluesky's AT Protocol has a public API for reading public data. Anything visible on the web is accessible without authentication—your social graph is designed to be portable and open.

## Rate limits

The tool adds small delays between requests to be respectful to the API. For accounts following 200+ people, expect it to take 2-5 minutes.

## License

MIT

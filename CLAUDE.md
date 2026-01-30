# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

bsky-orbit is a CLI tool that recommends Bluesky accounts to follow based on network graph analysis. It fetches a user's follows, analyzes who they follow, and ranks accounts by how many mutual connections exist.

## Build Commands

```bash
make build      # Build for local platform (outputs ./bsky-orbit)
make release    # Cross-compile for macOS, Linux, Windows (arm64/amd64)
make clean      # Remove compiled binaries
make tag v=X.Y.Z  # Create and push a git tag for release
```

## Running

```bash
./bsky-orbit <handle>        # e.g., ./bsky-orbit jasonlong.me
go run . <handle>            # Run without building
```

## Architecture

Single-file Go application (`main.go`) with no external dependencies beyond the standard library.

**Key components:**
- `apiGet()` - HTTP wrapper for Bluesky's public API with rate limiting
- `getProfile()` / `getAllFollows()` - API data fetchers with cursor-based pagination
- `main()` - Orchestrates the pipeline: fetch follows → analyze network → rank → output

**Data flow:**
1. Validate handle and get user DID
2. Fetch all accounts the user follows (paginated, 100 per request)
3. For each follow, fetch their follows and count occurrences
4. Filter out self and already-followed accounts
5. Rank by frequency, fetch profile details for top 30
6. Output: terminal table + JSON file + Markdown file

**API details:**
- Uses `public.api.bsky.app/xrpc` (no auth required)
- 50ms delay between requests (rate limiting)
- 15-second HTTP timeout per request

## Output Files

Running the tool generates:
- `bsky-orbit-{handle}.json` - Full recommendation data
- `bsky-orbit-{handle}.md` - Markdown table with Bluesky profile links

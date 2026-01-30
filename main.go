package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const apiBase = "https://public.api.bsky.app/xrpc"

type Profile struct {
	DID            string `json:"did"`
	Handle         string `json:"handle"`
	DisplayName    string `json:"displayName"`
	Description    string `json:"description"`
	FollowersCount int    `json:"followersCount"`
	FollowsCount   int    `json:"followsCount"`
}

type Follow struct {
	DID         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type FollowsResponse struct {
	Follows []Follow `json:"follows"`
	Cursor  string   `json:"cursor"`
}

type Recommendation struct {
	Rank           int    `json:"rank"`
	Handle         string `json:"handle"`
	DisplayName    string `json:"displayName"`
	Bio            string `json:"bio"`
	FollowersCount int    `json:"followersCount"`
	FollowedBy     int    `json:"followedByCount"`
	URL            string `json:"url"`
}

type Output struct {
	User            string           `json:"user"`
	Recommendations []Recommendation `json:"recommendations"`
}

var client = &http.Client{Timeout: 15 * time.Second}

func apiGet(endpoint string, params map[string]string) ([]byte, error) {
	u, _ := url.Parse(apiBase + "/" + endpoint)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", "bsky-orbit/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func getProfile(handle string) (*Profile, error) {
	data, err := apiGet("app.bsky.actor.getProfile", map[string]string{"actor": handle})
	if err != nil {
		return nil, err
	}
	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func getAllFollows(handle string) ([]Follow, error) {
	var follows []Follow
	cursor := ""

	for {
		params := map[string]string{"actor": handle, "limit": "100"}
		if cursor != "" {
			params["cursor"] = cursor
		}

		data, err := apiGet("app.bsky.graph.getFollows", params)
		if err != nil {
			break
		}

		var resp FollowsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			break
		}

		follows = append(follows, resp.Follows...)
		cursor = resp.Cursor
		if cursor == "" {
			break
		}
	}

	return follows, nil
}

func formatFollowers(count int) string {
	switch {
	case count >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(count)/1_000_000)
	case count >= 1_000:
		return fmt.Sprintf("%.1fK", float64(count)/1_000)
	default:
		return fmt.Sprintf("%d", count)
	}
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println(`bsky-orbit 🔭

Discover who to follow on Bluesky based on your network.

Usage:
    bsky-orbit <handle>

Examples:
    bsky-orbit jasonlong.me
    bsky-orbit @username.bsky.social

No API key required.`)
		os.Exit(1)
	}

	handle := strings.TrimPrefix(os.Args[1], "@")

	fmt.Printf("\n🔭 bsky-orbit: Analyzing @%s's network\n", handle)
	fmt.Println(strings.Repeat("=", 60))

	profile, err := getProfile(handle)
	if err != nil {
		fmt.Printf("Error: Could not find user @%s\n", handle)
		os.Exit(1)
	}
	myDID := profile.DID

	fmt.Println("\n📡 Fetching who you follow...")
	myFollows, err := getAllFollows(handle)
	if err != nil || len(myFollows) == 0 {
		fmt.Println("Error: Could not fetch follows")
		os.Exit(1)
	}
	fmt.Printf("   Found %d accounts\n", len(myFollows))

	followedDIDs := make(map[string]bool)
	for _, f := range myFollows {
		followedDIDs[f.DID] = true
	}

	fmt.Println("\n🔍 Analyzing who your follows follow...")
	fmt.Println("   This may take a few minutes for large networks")

	counts := make(map[string]int)
	info := make(map[string]Follow)

	for i, follow := range myFollows {
		fmt.Printf("\r   [%d/%d] %-45s", i+1, len(myFollows), truncate(follow.Handle, 40))

		theirFollows, _ := getAllFollows(follow.Handle)
		for _, f := range theirFollows {
			if f.DID == myDID || followedDIDs[f.DID] {
				continue
			}
			counts[f.DID]++
			if _, exists := info[f.DID]; !exists {
				info[f.DID] = f
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\n\n✨ Found %d accounts in your extended network\n", len(counts))

	type kv struct {
		DID   string
		Count int
	}
	var sorted []kv
	for did, count := range counts {
		sorted = append(sorted, kv{did, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})

	limit := 30
	if len(sorted) < limit {
		limit = len(sorted)
	}
	sorted = sorted[:limit]

	fmt.Printf("\n📊 Fetching details for top %d recommendations...\n", limit)

	var recommendations []Recommendation
	for i, kv := range sorted {
		f := info[kv.DID]
		fmt.Printf("\r   [%d/%d] @%-44s", i+1, limit, truncate(f.Handle, 40))

		rec := Recommendation{
			Rank:       i + 1,
			Handle:     f.Handle,
			FollowedBy: kv.Count,
			URL:        "https://bsky.app/profile/" + f.Handle,
		}

		if p, err := getProfile(f.Handle); err == nil {
			rec.DisplayName = p.DisplayName
			rec.Bio = truncate(p.Description, 100)
			rec.FollowersCount = p.FollowersCount
		}

		recommendations = append(recommendations, rec)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\n\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("🌟 TOP RECOMMENDATIONS FOR @%s\n", handle)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\n%-5s %-28s %-12s %s\n", "Rank", "Account", "Followers", "In Common")
	fmt.Println(strings.Repeat("-", 60))

	for _, rec := range recommendations {
		name := rec.DisplayName
		if name == "" {
			name = rec.Handle
		}
		fmt.Printf("%-5d %-28s %-12s %d\n", rec.Rank, truncate(name, 26), formatFollowers(rec.FollowersCount), rec.FollowedBy)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("\n'In Common' = how many of your follows also follow this account")

	jsonFile := fmt.Sprintf("bsky-orbit-%s.json", handle)
	output := Output{User: handle, Recommendations: recommendations}
	jsonData, _ := json.MarshalIndent(output, "", "  ")
	os.WriteFile(jsonFile, jsonData, 0644)

	mdFile := fmt.Sprintf("bsky-orbit-%s.md", handle)
	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Bluesky Follow Recommendations for @%s\n\n", handle))
	md.WriteString("Accounts you might want to follow, ranked by how many of your follows also follow them.\n\n")
	md.WriteString("| Rank | Account | Followers | In Common | Bio |\n")
	md.WriteString("|------|---------|-----------|-----------|-----|\n")
	for _, rec := range recommendations {
		name := rec.DisplayName
		if name == "" {
			name = rec.Handle
		}
		bio := strings.ReplaceAll(rec.Bio, "|", "\\|")
		md.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %d | %s |\n",
			rec.Rank, name, rec.URL, formatFollowers(rec.FollowersCount), rec.FollowedBy, bio))
	}
	os.WriteFile(mdFile, []byte(md.String()), 0644)

	fmt.Printf("\n📁 Saved: %s\n", jsonFile)
	fmt.Printf("📁 Saved: %s\n", mdFile)
	fmt.Println("\n🚀 Done! Happy following.\n")
}

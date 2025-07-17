package channelscrapper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	slerr "github.com/defany/slogger/pkg/err"
)

const fetchLastPostTTL = 5 * time.Minute

type fetchLastPostCacheEntry struct {
	text     string
	postTime time.Time
	fetched  time.Time
}

var (
	fetchLastPostCacheMu sync.RWMutex
	fetchLastPostCache   = make(map[string]fetchLastPostCacheEntry)
)

func FetchLastPost(ctx context.Context, channelName string) (string, time.Time, error) {
	fetchLastPostCacheMu.RLock()
	if e, ok := fetchLastPostCache[channelName]; ok && time.Since(e.fetched) < fetchLastPostTTL {
		fetchLastPostCacheMu.RUnlock()
		return e.text, e.postTime, nil
	}
	fetchLastPostCacheMu.RUnlock()

	url := fmt.Sprintf("https://t.me/s/%s", channelName)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", time.Time{}, slerr.WithSource(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, slerr.WithSource(errors.New("status code is not 200"))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", time.Time{}, slerr.WithSource(err)
	}

	messages := doc.Find(".tgme_widget_message_wrap")
	var (
		msgSel *goquery.Selection
		text   string
	)
	for i := messages.Length() - 1; i >= 0; i-- {
		sel := messages.Eq(i)
		txt := strings.TrimSpace(sel.Find(".tgme_widget_message_text").Text())
		if txt == "" {
			continue
		}
		msgSel = sel
		text = txt
		break
	}

	if msgSel == nil {
		return "", time.Time{}, slerr.WithSource(errors.New("no text post found"))
	}

	timeAttr, ok := msgSel.Find(".tgme_widget_message_date time").Attr("datetime")
	if !ok {
		return "", time.Time{}, slerr.WithSource(errors.New("datetime attribute not found"))
	}

	parsed, err := time.Parse(time.RFC3339, timeAttr)
	if err != nil {
		return "", time.Time{}, slerr.WithSource(err)
	}

	loc := time.Local
	now := time.Now().In(loc)
	postTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		parsed.In(loc).Hour(), parsed.In(loc).Minute(), 0, 0,
		loc,
	)

	fetchLastPostCacheMu.Lock()
	fetchLastPostCache[channelName] = fetchLastPostCacheEntry{
		text:     text,
		postTime: postTime,
		fetched:  time.Now(),
	}
	fetchLastPostCacheMu.Unlock()

	return text, postTime, nil
}

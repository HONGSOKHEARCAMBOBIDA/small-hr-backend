package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	reAtSign = regexp.MustCompile(`@(-?\d+\.\d+),(-?\d+\.\d+)`)
	reDataLL = regexp.MustCompile(`!3d(-?\d+\.\d+)!4d(-?\d+\.\d+)`)
)

// ExtractLatLngFromGoogleMapsURL accepts a Google Maps URL in one of the
// common formats and returns latitude/longitude as strings.
//
// Supported formats:
//
//	https://maps.google.com/maps?q=12.806047,103.513508&ll=12.806047,103.513508&z=16
//	https://www.google.com/maps/@12.806047,103.513508,16z
//	https://www.google.com/maps/place/Some+Place/@12.806047,103.513508,17z/data=...!3d12.806047!4d103.513508
//	https://maps.app.goo.gl/xxxxx   (short link, resolved via redirect)
//	https://goo.gl/maps/xxxxx       (short link, resolved via redirect)
func ExtractLatLngFromGoogleMapsURL(raw string) (lat, lng string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("map link is empty")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("invalid url: %w", err)
	}

	// Short links don't carry coordinates in the URL itself — resolve the
	// redirect first to get the real maps URL.
	if strings.Contains(u.Host, "goo.gl") {
		resolved, rerr := resolveRedirectLocation(raw)
		if rerr != nil {
			return "", "", fmt.Errorf("could not resolve short link: %w", rerr)
		}
		raw = resolved
		if u, err = url.Parse(raw); err != nil {
			return "", "", fmt.Errorf("invalid resolved url: %w", err)
		}
	}

	// 1) ?q=lat,lng or ?ll=lat,lng (your example format)
	q := u.Query()
	for _, key := range []string{"q", "ll", "query", "destination"} {
		if v := q.Get(key); v != "" {
			if lt, lg, ok := splitLatLng(v); ok {
				return lt, lg, nil
			}
		}
	}

	// 2) .../@lat,lng,zoomz...
	if m := reAtSign.FindStringSubmatch(raw); len(m) == 3 {
		return m[1], m[2], nil
	}

	// 3) ...!3dlat!4dlng... (shows up in "place" links)
	if m := reDataLL.FindStringSubmatch(raw); len(m) == 3 {
		return m[1], m[2], nil
	}

	return "", "", errors.New("could not find coordinates in the provided map link")
}

func splitLatLng(s string) (lat, lng string, ok bool) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	latF, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lngF, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return "", "", false
	}
	return strconv.FormatFloat(latF, 'f', -1, 64),
		strconv.FormatFloat(lngF, 'f', -1, 64), true
}

// resolveRedirectLocation follows a single redirect hop and returns the
// Location header value, without downloading the full destination page.
func resolveRedirectLocation(shortURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(shortURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	if loc == "" {
		return shortURL, nil // already a full URL, nothing to resolve
	}
	return loc, nil
}

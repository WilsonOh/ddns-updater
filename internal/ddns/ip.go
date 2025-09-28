package ddns

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

const CLOUDFLARE_TRACE_URL = "https://www.cloudflare.com/cdn-cgi/trace"

func GetPublicIPAddress() (string, error) {
	resp, err := http.Get(CLOUDFLARE_TRACE_URL)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	kvPairs := strings.SplitSeq(string(body), "\n")
	for pair := range kvPairs {
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, "=")
		key := parts[0]
		value := parts[1]
		if strings.EqualFold(key, "ip") && value != "" {
			return value, nil
		}
	}
	return "", errors.New("ip key not found in trace response")
}

package gopify

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

type Pagination struct {
	Previous string
	Next     string
}

func extractPagination(linkHeader string) (*Pagination, error) {
	linkRegex, err := regexp.Compile(`<([^<]+)>; rel="(previous|next)"`)
	if err != nil {
		return nil, err
	}
	pagination := new(Pagination)
	if linkHeader == "" {
		return pagination, nil
	}
	for _, link := range strings.Split(linkHeader, ",") {
		matches := linkRegex.FindStringSubmatch(link)
		if len(matches) != 3 {
			return nil, errors.New("Invalid pagination link header")
		}
		u, err := url.Parse(matches[1])
		if err != nil {
			return nil, err
		}
		cursor := u.Query().Get("page_info")
		if matches[2] == "next" {
			pagination.Next = cursor
		} else {
			pagination.Previous = cursor
		}
	}
	return pagination, nil
}

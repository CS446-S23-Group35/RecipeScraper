package linksource

import "golang.org/x/net/html"

// LinkPage is a struct that contains a list of links and the link to
// the next page link.
type LinkPage struct {
	Links    []string
	NextPage string
}

// LinkSource is an interface for parsing links from an HTML home page.
// It takes a HTML node and parses it into a list of links and the link
// to the next page to parse.
type LinkSource interface {
	GetLinks(*html.Node) (*LinkPage, error)
}

package linksource

import (
	"fmt"
	"strings"

	css "github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

const (
	foodNetworkLinkSelector     = "li.m-PromoList__a-ListItem a[href]"
	foodNetworkNextPageSelector = "a[href].o-Pagination__a-NextButton:not(.is-Disabled)"
	foodNetworkCategorySelector = "h3.o-Capsule__a-Headline"
	// foodNetworkNextCategorySelector = "ul.o-IndexPagination__m-List"
	foodNetworkBaseLink = "https://www.foodnetwork.com/recipes/recipes-a-z/"
)

type FoodNetworkLinkSource struct {
	linkSelector     css.Selector
	nextPageSelector css.Selector
	categorySelector css.Selector
}

func NewFoodnetworkLinkSource() *FoodNetworkLinkSource {
	return &FoodNetworkLinkSource{
		linkSelector:     css.MustCompile(foodNetworkLinkSelector),
		nextPageSelector: css.MustCompile(foodNetworkNextPageSelector),
		categorySelector: css.MustCompile(foodNetworkCategorySelector),
	}
}

func (f FoodNetworkLinkSource) GetLinks(node *html.Node) (*LinkPage, error) {
	// Get all links from the page
	linkNodes := f.linkSelector.MatchAll(node)
	links := make([]string, len(linkNodes))
	for i, linkNode := range linkNodes {
		for _, attr := range linkNode.Attr {
			if attr.Key == "href" {
				links[i] = "https:" + attr.Val
			}
		}
	}

	// Try to find the next page link
	nextPageNode := f.nextPageSelector.MatchFirst(node)
	if nextPageNode != nil {
		for _, attr := range nextPageNode.Attr {
			if attr.Key == "href" {
				return &LinkPage{
					Links:    links,
					NextPage: "https:" + attr.Val,
				}, nil
			}
		}
	}

	// If there is no next page link, try to make one from the category
	categoryNode := f.categorySelector.MatchFirst(node)
	category := ""
	if categoryNode != nil {
		for _, attr := range categoryNode.Attr {
			if attr.Key == "id" {
				category = attr.Val
			}
		}
	}

	if category != "" {
		// Try to get next category, if there is none return ""
		nextPage := nextCategory(category)
		if nextPage != "" {
			nextPage = foodNetworkBaseLink + nextPage
		}

		return &LinkPage{
			Links:    links,
			NextPage: nextPage,
		}, nil
	}

	return nil, fmt.Errorf("did not find category or next page link")
}

// NextCategory returns the next category to parse.
func nextCategory(cat string) string {
	switch strings.ToLower(cat) {
	case "":
		return "123"
	case "123":
		return "a"
	case "w":
		return "xyz"
	case "xyz":
		return ""
	default:
		bytes := []byte(cat)
		for i := 0; i < len(bytes); i++ {
			bytes[i]++
		}
		return string(bytes)
	}
}

package linksource

import (
	"fmt"
	"strings"

	css "github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

const (
	foodNetworkLinkSelector     = "li.m-PromoList__a-ListItem a[href]"
	foodNetworkNextPageSelector = "a[href].o-Pagination__a-NextButton"
	foodNetworkCategorySelector = "h3.o-Capsule__a-Headline"
	// foodNetworkNextCategorySelector = "ul.o-IndexPagination__m-List"
	foodNetworkBaseLink = "https://www.foodnetwork.com/recipes/a-z/"
)

type FoodNetworkLinkSource struct {
	linkSelector     css.Selector
	nextPageSelector css.Selector
	categorySelector css.Selector
}

func NewFoodNetworkLinkSource() *FoodNetworkLinkSource {
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
				links[i] = attr.Val
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
					NextPage: attr.Val,
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
		return &LinkPage{
			Links:    links,
			NextPage: foodNetworkBaseLink + strings.ToLower(nextCategory(category)),
		}, nil
	}

	return nil, fmt.Errorf("did not find category or next page link")
}

// NextCategory returns the next category to parse.
func nextCategory(cat string) string {
	switch cat {
	case "":
		return "123"
	case "123":
		return "A"
	case "W":
		return "XYZ"
	case "XYZ":
		return ""
	default:
		bytes := []byte(cat)
		for i := 0; i < len(bytes); i++ {
			bytes[i]++
		}
		return string(bytes)
	}
}

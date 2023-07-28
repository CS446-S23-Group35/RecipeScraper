package scraper

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/CS446-S23-Group35/RecipeScraper/pkg/linksource"
	"github.com/CS446-S23-Group35/RecipeScraper/pkg/parser"
	"github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"
	"golang.org/x/net/html"
	yaml "gopkg.in/yaml.v3"
)

type Scraper struct {
	linkSource linksource.LinkSource
	parser     parser.Parser
	writer     io.WriteCloser
	startLink  string
	onlyLinks  bool
}

func NewScraper(cfg Config) *Scraper {
	outputFile, err := os.Create(cfg.OutputPath)
	if err != nil {
		panic(err)
	}

	switch cfg.SourceType {
	case "foodnetwork":
		return &Scraper{
			linkSource: linksource.NewFoodnetworkLinkSource(),
			parser:     parser.NewFoodnetworkParser(),
			writer:     outputFile,
			startLink:  cfg.StartLink,
			onlyLinks:  cfg.OnlyLinks,
		}
	}
	return nil
}

func (s *Scraper) Scrape(ctx context.Context) error {
	defer s.writer.Close()

	linkFile, err := os.Create("links.tmp")
	if err != nil {
		return fmt.Errorf("could not create links temp file: %w", err)
	}
	defer linkFile.Close()

	links := make([]string, 0, 100000)

	curLink := s.startLink
	for {
		log.Println("Doing page: " + curLink)
		linkPage, err := s.scrapeForLink(ctx, curLink)
		if err != nil {
			return fmt.Errorf("error scraping for link: %w", err)
		}

		links = append(links, linkPage.Links...)

		if linkPage.NextPage == "" {
			break
		}
		curLink = linkPage.NextPage
	}

	linkOut := strings.Join(links, "\n") + "\n"
	linkFile.WriteString(linkOut)

	if s.onlyLinks {
		return nil
	}

	log.Println("Starting scraping recipes")

	for _, link := range links {
		err := s.scrapeRecipe(ctx, link)
		if err != nil {
			msg := fmt.Sprintf("error scraping recipe at link %s: %s", link, err.Error())
			log.Println(msg)
		}
	}

	log.Println("Finishing scraping with Success!")
	return nil
}

func (s *Scraper) ScrapeFromLinksFile(ctx context.Context, filepath string) error {
	linkFile, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("could not open links file: %w", err)
	}
	defer linkFile.Close()

	linkBytes, err := io.ReadAll(linkFile)
	if err != nil {
		return fmt.Errorf("could not read links file: %w", err)
	}

	links := strings.Split(strings.Trim(string(linkBytes), "\n"), "\n")
	shuffle(links)
	return s.scrapeRecipes(ctx, links)
}

func (s *Scraper) scrapeRecipes(ctx context.Context, links []string) error {
	defer s.writer.Close()
	defer s.writeLinks(&links)

	for _, link := range links {
		err := s.scrapeRecipe(ctx, link)
		if err != nil {
			msg := fmt.Sprintf("error scraping recipe at link %s: %s", link, err.Error())
			log.Println(msg)
		}
	}
	return nil
}

func (s *Scraper) writeLinks(links *[]string) error {
	file, err := os.Create("remainingLinks.txt")
	if err != nil {
		log.Println("Error writing remaining links, printing instead")
		for _, link := range *links {
			log.Println(link)
		}
		return fmt.Errorf("could not create remaining links file: %w", err)
	}
	defer file.Close()

	linkOut := strings.Join(*links, "\n") + "\n"
	_, err = file.WriteString(linkOut)
	if err != nil {
		log.Println("Error writing remaining links, printing instead")
		for _, link := range *links {
			log.Println(link)
		}
		return fmt.Errorf("could not write remaining links file: %w", err)
	}

	return nil
}

func (s *Scraper) scrapeRecipe(ctx context.Context, link string) error {
	page, err := s.makeRequest(link)
	if err != nil {
		return fmt.Errorf("error scraping recipe: %w", err)
	}

	node, err := html.Parse(page.Body)
	if err != nil {
		return fmt.Errorf("error parsing recipe: %w", err)
	}

	rawRecipe, err := s.parser.ParseRecipe(node)
	if err != nil {
		return fmt.Errorf("error parsing recipe: %w", err)
	}
	rawRecipe.Metadata.SourceURL = link

	listFmt := make([]recipe.RawRecipe, 1)
	listFmt[0] = *rawRecipe

	yamlBytes, err := yaml.Marshal(listFmt)
	if err != nil {
		return fmt.Errorf("error marshalling recipe: %w", err)
	}

	_, err = s.writer.Write(append(yamlBytes, '\n'))
	if err != nil {
		return fmt.Errorf("error writing recipe: %w", err)
	}

	return nil
}

func (s *Scraper) scrapeForLink(ctx context.Context, link string) (*linksource.LinkPage, error) {
	page, err := s.makeRequest(link)
	log.Println("Link Page Response: ", page.Status)
	if err != nil {
		return nil, fmt.Errorf("error scraping for links: %w", err)
	}

	node, err := html.Parse(page.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing for links: %w", err)
	}

	links, err := s.linkSource.GetLinks(node)
	if err != nil {
		return nil, fmt.Errorf("error parsing for links: %w", err)
	}
	return links, nil
}

func (s *Scraper) makeRequest(link string) (*http.Response, error) {
	timeToSleep := int64(rand.Float64()*10) + 2
	time.Sleep(time.Duration(timeToSleep) * time.Second)
	return http.Get(link)
}

func shuffle(links []string) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(links), func(i, j int) { links[i], links[j] = links[j], links[i] })
}

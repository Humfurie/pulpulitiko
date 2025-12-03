package handlers

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type RSSHandler struct {
	articleService *services.ArticleService
	siteURL        string
}

func NewRSSHandler(articleService *services.ArticleService, siteURL string) *RSSHandler {
	return &RSSHandler{
		articleService: articleService,
		siteURL:        siteURL,
	}
}

// RSS 2.0 structures
type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate"`
	AtomLink      AtomLink  `xml:"atom:link"`
	Items         []RSSItem `xml:"item"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author,omitempty"`
	Category    string `xml:"category,omitempty"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

// GET /rss or /feed
func (h *RSSHandler) Feed(w http.ResponseWriter, r *http.Request) {
	// Get latest published articles
	status := models.ArticleStatusPublished
	filter := &models.ArticleFilter{
		Status: &status,
	}

	articles, err := h.articleService.List(r.Context(), filter, 1, 20)
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	}

	// Build RSS items
	items := make([]RSSItem, 0, len(articles.Articles))
	for _, article := range articles.Articles {
		description := ""
		if article.Summary != nil {
			description = *article.Summary
		}

		pubDate := ""
		if article.PublishedAt != nil {
			pubDate = article.PublishedAt.Format(time.RFC1123Z)
		}

		author := ""
		if article.AuthorName != nil {
			author = *article.AuthorName
		}

		category := ""
		if article.CategoryName != nil {
			category = *article.CategoryName
		}

		items = append(items, RSSItem{
			Title:       article.Title,
			Link:        h.siteURL + "/article/" + article.Slug,
			Description: description,
			Author:      author,
			Category:    category,
			GUID:        h.siteURL + "/article/" + article.Slug,
			PubDate:     pubDate,
		})
	}

	rss := RSS{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: RSSChannel{
			Title:         "Pulpulitiko - Philippine Politics News",
			Link:          h.siteURL,
			Description:   "Your trusted source for Philippine political news and commentary",
			Language:      "en-ph",
			LastBuildDate: time.Now().Format(time.RFC1123Z),
			AtomLink: AtomLink{
				Href: h.siteURL + "/rss",
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=900") // 15 minutes cache

	_, _ = w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(rss); err != nil {
		http.Error(w, "Failed to encode RSS", http.StatusInternalServerError)
		return
	}
}

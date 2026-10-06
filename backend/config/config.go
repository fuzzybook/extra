package config

import "strings"

type Category struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Config struct {
	Categories            []Category `json:"categories"`
	BroadcasterAttributes []string   `json:"broadcaster_attributes"`
	Availability          []string   `json:"availability"`
	BroadcasterServices   []string   `json:"broadcaster_services"`
}

var SiteConfig = Config{
	Categories: []Category{
		{
			Name:  "General",
			Color: "#63f18eff",
		},
		{
			Name:  "Music",
			Color: "#f16365ff",
		},
		{
			Name:  "Beauty",
			Color: "#6366f1",
		},
		{
			Name:  "test",
			Color: "#ecf163ff",
		},
	},
	BroadcasterAttributes: []string{
		"Genere Musicale",
		"Strumento",
		"Anni di Esperienza",
		"Città Base",
		"Instagram",
		"Spotify",
		"YouTube",
		"SoundCloud",
		"Attrezzatura DJ",
		"prova attributo",
	},
	Availability: []string{
		"incall",
		"autcall",
		"incall & autcall",
	},
	BroadcasterServices: []string{
		"Live DJ",
		"Produzioni",
		"Vocalist",
		"Musicista",
	},
}

func GetBroadcasterAttributes() []string {
	if SiteConfig.BroadcasterAttributes == nil {
		return []string{}
	}
	return SiteConfig.BroadcasterAttributes
}

func GetBroadcasterServices() []string {
	if SiteConfig.BroadcasterServices == nil {
		return []string{}
	}
	return SiteConfig.BroadcasterServices
}

func GetAvailabilityOptions() []string {
	if SiteConfig.Availability == nil {
		return []string{}
	}
	return SiteConfig.Availability
}

func GetCategory(name string) *Category {
	for _, c := range SiteConfig.Categories {
		if strings.EqualFold(c.Name, name) {
			return &c
		}
	}
	return nil
}

func GetCategoryColor(name string) string {
	if c := GetCategory(name); c != nil && c.Color != "" {
		return c.Color
	}
	return "#6366f1"
}

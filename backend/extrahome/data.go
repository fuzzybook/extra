package extrahome

import "time"

type SiteConfig struct {
	ExtraSite   ExtraSite
	StartAt     time.Time
	Description string
}

type ExtraSite struct {
	NavItems            []NavItem
	HomePageData        HomePageData
	BroadcasterPageData BroadcasterPageData
	ContactPageData     ContactPageData
	Carusel             Carusel
}

type NavItem struct {
	Label string
	Color string
	URL   string
}

type HomePageData struct {
	Title              string
	Lang               string
	HeroTitle          string
	Profiles           []Profile
	Promo              Promo
	ListOfBroadcasters ListOfBroadcaster
	Carusel            Carusel
	Year               int
}

type BroadcasterPageData struct {
	Title       string
	Lang        string
	Year        int
	PromoRight1 Promo
	PromoRight2 Promo
	Profile     BroadcasterProfile
}

type ContactPageData struct {
	Title     string
	Lang      string
	HeroTitle string
	Email     string
	Phone     string
	Address   string
	Hours     string
	Message   string
	Year      int
}

type Carusel struct {
	Items []CaruselItem
}

type CaruselItem struct {
	Title       string
	Description string
	ImageURL    string
	URL         string
}

type Promo struct {
	Left   PromoItem
	Right1 PromoItem
	Right2 PromoItem
}

type PromoItem struct {
	Title       string
	Description string
	Content     []PromoContent
}

type PromoContent struct {
	Width          int
	Height         int
	MobileOnly     bool
	MobileImageURL string
	NewWindow      bool
	Title          string
	Description    string
	ImageURL       string
	URL            string
}

type ListOfBroadcaster struct {
	Title string
	Items []ListOfBroadcasterItem
}

type ListOfBroadcasterItem struct {
	Title       string
	Category    string
	Description string
	ImageURL    string
	URL         string
}

type Profile struct {
	Name         string
	ProfilePhoto ProfilePhoto
	Category     string
	Description  string
	URL          string
}

type ProfilePhoto struct {
	ImageURL string
	Alt      string
}

type ProfileHours struct {
	Day   string
	Hours string
}

type ProfileAttribute struct {
	Label string
	Value string
}

type BroadcasterProfile struct {
	ID            int64
	Name          string
	Category      string
	CategoryColor string
	Certified     bool
	Introduction  []string
	Photos        []ProfilePhoto
	Hours         []ProfileHours
	Availability  string
	Phone         string
	PhoneNumber   string
	Attributes    []ProfileAttribute
	Languages     []string
	Services      []string
}

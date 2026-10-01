package extrahome

import "time"

type Head struct {
	Title   string
	Meta    []HeadMeta
	Links   []HeadLink
	Langs   []HeadLang
	Scripts []HeadScript
}
type HeadMeta struct {
	Name    string
	Content string
}
type HeadLink struct {
	Rel  string
	Href string
}
type HeadScript struct {
	Defer      string
	DataDomain string
	Src        string
}
type HeadLang struct {
	Rel      string
	Hreflang string
	Href     string
}

type SiteLangs map[string][]struct {
	Name string
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

type ListOfBroadcaster struct {
	Title        string
	Broadcasters []int64
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

const (
	PromoLeft = iota
	PromoRight1
	PromoRight2
)

type Promo struct {
	ID         int64
	Position   int // Promo position: PromoLeft, PromoRight1, PromoRight2
	Item       []PromoItem
	ActiveFrom time.Time
}

type PromoItem struct {
	Title   string
	Content []PromoContent
}

type PromoContent struct {
	Width          int
	Height         int
	MobileOnly     bool
	MobileImageURL string
	NewWindow      bool
	ImageURL       string
	URL            string
}

type NavItem struct {
	Label string
	Color string
	URL   string
}

type HomePageData struct {
	Title              string
	SiteLangs          SiteLangs
	PromoLeft          Promo // Promo position: PromoLeft id
	PromoRight1        Promo // Promo position: PromoRight1 id
	PromoRight2        Promo // Promo position: PromoRight2 id
	ListOfBroadcasters ListOfBroadcaster
	Carusel            Carusel
	Year               int
}

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

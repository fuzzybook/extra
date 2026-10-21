package advertising

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	authsvc "encore.app/auth"
	"encore.app/validate"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

type PromoPosition int

const (
	PromoLeft   PromoPosition = 0
	PromoRight1 PromoPosition = 1
	PromoRight2 PromoPosition = 2
)

type PromoLeftSections string

const (
	PromoLeftWeek        PromoLeftSections = "week"
	PromoLeftRaccomended PromoLeftSections = "raccomended"
	PromoLeftDiscover    PromoLeftSections = "discover"
)

type PromoImage struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Proportion string `json:"proportion"`
}

var PromoLeftImageSize = PromoImage{
	Width:      230,
	Height:     0,
	Proportion: "2:3",
}

var PromoRight1ImageSize = PromoImage{
	Width:      230,
	Height:     0,
	Proportion: "2:3",
}

var PromoRight2ImageSize = PromoImage{
	Width:      230,
	Height:     345,
	Proportion: "",
}

type PromoRight1Sections string

const (
	PromoRight1Show    PromoRight1Sections = "show"
	PromoRight1Service PromoRight1Sections = "service"
)

type PromoRight2Sections string

const (
	PromoRight2Top      PromoRight2Sections = "top"
	PromoRight2Selected PromoRight2Sections = "selected"
)

type PromoContent struct {
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	MobileOnly     bool   `json:"mobile_only"`
	MobileImageURL string `json:"mobile_image_url"`
	NewWindow      bool   `json:"new_window"`
	ImageURL       string `json:"image_url"`
	URL            string `json:"url"`
}

type PromoItem struct {
	Title   string         `json:"title"`
	Content []PromoContent `json:"content"`
}

type Promo struct {
	ID         int64         `json:"id"`
	Position   PromoPosition `json:"position"`
	Section    string        `json:"section"`
	Item       []PromoItem   `json:"item"`
	Year       int           `json:"year"`
	Weeks      []int         `json:"weeks"`
	ActiveFrom time.Time     `json:"active_from"`
	CreatedAt  time.Time     `json:"created_at"`
}

type CreatePromoParams struct {
	Position   PromoPosition `json:"position"`
	Section    string        `json:"section"`
	Item       []PromoItem   `json:"item"`
	Year       int           `json:"year,omitempty"`
	Weeks      []int         `json:"weeks,omitempty"`
	ActiveFrom *time.Time    `json:"active_from,omitempty"`
}

func validatePromoFields(position PromoPosition, section string) error {
	if position < PromoLeft || position > PromoRight2 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "posizione promo non valida, ammessi solo 0 (PromoLeft), 1 (PromoRight1), 2 (PromoRight2)"}
	}
	switch position {
	case PromoLeft:
		if section != string(PromoLeftWeek) && section != string(PromoLeftRaccomended) && section != string(PromoLeftDiscover) {
			return &errs.Error{Code: errs.InvalidArgument, Message: "sezione non valida per PromoLeft, ammessi solo: 'week', 'raccomended', 'discover'"}
		}
	case PromoRight1:
		if section != string(PromoRight1Show) && section != string(PromoRight1Service) {
			return &errs.Error{Code: errs.InvalidArgument, Message: "sezione non valida per PromoRight1, ammessi solo: 'show', 'service'"}
		}
	case PromoRight2:
		if section != string(PromoRight2Top) && section != string(PromoRight2Selected) {
			return &errs.Error{Code: errs.InvalidArgument, Message: "sezione non valida per PromoRight2, ammessi solo: 'top', 'selected'"}
		}
	}
	return nil
}

// validatePromoItems accetta solo URL http/https: i link e le immagini delle
// promo finiscono in href/src del sito pubblico.
func validatePromoItems(items []PromoItem) error {
	for _, item := range items {
		for _, c := range item.Content {
			for _, u := range []string{c.URL, c.ImageURL, c.MobileImageURL} {
				if err := validate.HTTPURL(u); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validatePromoWeeks(weeks []int) error {
	for _, w := range weeks {
		if w < 1 || w > 53 {
			return &errs.Error{Code: errs.InvalidArgument, Message: "numero settimana non valido, deve essere compreso tra 1 e 53"}
		}
	}
	return nil
}

func (p *CreatePromoParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "i parametri di creazione sono obbligatori"}
	}
	if err := validatePromoFields(p.Position, p.Section); err != nil {
		return err
	}
	if p.Item == nil {
		p.Item = []PromoItem{}
	}
	if err := validatePromoItems(p.Item); err != nil {
		return err
	}
	if p.Weeks == nil {
		p.Weeks = []int{}
	}
	if err := validatePromoWeeks(p.Weeks); err != nil {
		return err
	}
	if p.Year == 0 {
		p.Year = time.Now().Year()
	}
	return nil
}

type UpdatePromoParams struct {
	Position   PromoPosition `json:"position"`
	Section    string        `json:"section"`
	Item       []PromoItem   `json:"item"`
	Year       int           `json:"year,omitempty"`
	Weeks      []int         `json:"weeks,omitempty"`
	ActiveFrom *time.Time    `json:"active_from,omitempty"`
}

func (p *UpdatePromoParams) Validate() error {
	if p == nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "i parametri di aggiornamento sono obbligatori"}
	}
	if err := validatePromoFields(p.Position, p.Section); err != nil {
		return err
	}
	if p.Item == nil {
		p.Item = []PromoItem{}
	}
	if err := validatePromoItems(p.Item); err != nil {
		return err
	}
	if p.Weeks == nil {
		p.Weeks = []int{}
	}
	if err := validatePromoWeeks(p.Weeks); err != nil {
		return err
	}
	if p.Year == 0 {
		p.Year = time.Now().Year()
	}
	return nil
}

type ListPromoResponse struct {
	Promos []*Promo `json:"promos"`
}

type ListPromotionsResponse struct {
	Promotions []*Promo `json:"promotions"`
}

func queryPromos(ctx context.Context) ([]*Promo, error) {
	rows, err := db.Query(ctx, `
		SELECT id, position, section, item, year, weeks, active_from, created_at
		FROM promo
		ORDER BY active_from DESC, id DESC
	`)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante il recupero delle promozioni")
	}
	defer rows.Close()

	var promos []*Promo
	for rows.Next() {
		var (
			p          Promo
			itemBytes  []byte
			weeksBytes []byte
		)
		if err := rows.Scan(&p.ID, &p.Position, &p.Section, &itemBytes, &p.Year, &weeksBytes, &p.ActiveFrom, &p.CreatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "errore durante la lettura del record promo")
		}
		if len(itemBytes) > 0 {
			if err := json.Unmarshal(itemBytes, &p.Item); err != nil {
				p.Item = []PromoItem{}
			}
		} else {
			p.Item = []PromoItem{}
		}
		if len(weeksBytes) > 0 {
			if err := json.Unmarshal(weeksBytes, &p.Weeks); err != nil {
				p.Weeks = []int{}
			}
		} else {
			p.Weeks = []int{}
		}
		promos = append(promos, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "errore durante l'iterazione delle promozioni")
	}

	if promos == nil {
		promos = []*Promo{}
	}

	return promos, nil
}

// ListPromos ritorna la lista completa di tutte le promozioni.
//
//encore:api public method=GET path=/promo
func ListPromos(ctx context.Context) (*ListPromoResponse, error) {
	promos, err := queryPromos(ctx)
	if err != nil {
		return nil, err
	}
	return &ListPromoResponse{Promos: promos}, nil
}

// ListPromotions ritorna la lista completa delle promozioni arricchite con le settimane lavorative attive.
//
//encore:api public method=GET path=/promotions
func ListPromotions(ctx context.Context) (*ListPromotionsResponse, error) {
	promos, err := queryPromos(ctx)
	if err != nil {
		return nil, err
	}
	return &ListPromotionsResponse{Promotions: promos}, nil
}

// CreatePromo inserisce una nuova promozione nel database. Solo admin.
//
//encore:api auth method=POST path=/promo
func CreatePromo(ctx context.Context, p *CreatePromoParams) (*Promo, error) {
	if err := authsvc.RequireAdmin(); err != nil {
		return nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}

	itemJSON, err := json.Marshal(p.Item)
	if err != nil {
		return nil, errs.WrapCode(err, errs.InvalidArgument, "formato item non valido")
	}

	weeksJSON, err := json.Marshal(p.Weeks)
	if err != nil {
		return nil, errs.WrapCode(err, errs.InvalidArgument, "formato settimane non valido")
	}

	activeFrom := time.Now()
	if p.ActiveFrom != nil && !p.ActiveFrom.IsZero() {
		activeFrom = *p.ActiveFrom
	}

	var promo Promo
	promo.Position = p.Position
	promo.Section = p.Section
	promo.Item = p.Item
	promo.Year = p.Year
	promo.Weeks = p.Weeks
	promo.ActiveFrom = activeFrom

	err = db.QueryRow(ctx, `
		INSERT INTO promo (position, section, item, year, weeks, active_from)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, p.Position, p.Section, itemJSON, p.Year, weeksJSON, activeFrom).Scan(&promo.ID, &promo.CreatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile creare la promozione")
	}

	return &promo, nil
}

// UpdatePromo aggiorna una promozione esistente nel database. Solo admin.
//
//encore:api auth method=PUT path=/promo/:id
func UpdatePromo(ctx context.Context, id int64, p *UpdatePromoParams) (*Promo, error) {
	if err := authsvc.RequireAdmin(); err != nil {
		return nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}

	itemJSON, err := json.Marshal(p.Item)
	if err != nil {
		return nil, errs.WrapCode(err, errs.InvalidArgument, "formato item non valido")
	}

	weeksJSON, err := json.Marshal(p.Weeks)
	if err != nil {
		return nil, errs.WrapCode(err, errs.InvalidArgument, "formato settimane non valido")
	}

	var promo Promo
	promo.ID = id
	promo.Position = p.Position
	promo.Section = p.Section
	promo.Item = p.Item
	promo.Year = p.Year
	promo.Weeks = p.Weeks

	query := `
		UPDATE promo
		SET position = $1, section = $2, item = $3, year = $4, weeks = $5,
		    active_from = COALESCE($6, active_from)
		WHERE id = $7
		RETURNING active_from, created_at
	`
	err = db.QueryRow(ctx, query, p.Position, p.Section, itemJSON, p.Year, weeksJSON, p.ActiveFrom, id).Scan(&promo.ActiveFrom, &promo.CreatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "promozione non trovata"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "impossibile aggiornare la promozione")
	}

	return &promo, nil
}

// PromoLeftGroup raggruppa le promozioni di PromoLeft per sezione ordinata.
type PromoLeftGroup struct {
	Week        []*Promo `json:"week"`
	Raccomended []*Promo `json:"raccomended"`
	Discover    []*Promo `json:"discover"`
}

// PromoRight1Group raggruppa le promozioni di PromoRight1 per sezione ordinata.
type PromoRight1Group struct {
	Show    []*Promo `json:"show"`
	Service []*Promo `json:"service"`
}

// PromoRight2Group raggruppa le promozioni di PromoRight2 per sezione ordinata.
type PromoRight2Group struct {
	Top      []*Promo `json:"top"`
	Selected []*Promo `json:"selected"`
}

// WeekPromoBySection contiene i tre tipi di promo suddivisi per sezioni ordinate.
type WeekPromoBySection struct {
	PromoLeft   PromoLeftGroup   `json:"promo_left"`
	PromoRight1 PromoRight1Group `json:"promo_right1"`
	PromoRight2 PromoRight2Group `json:"promo_right2"`
}

// WeekPromoResponse rappresenta il payload JSON restituito dall'endpoint GetWeekPromo.
type WeekPromoResponse struct {
	Week        int                 `json:"week"`
	Year        int                 `json:"year"`
	PromoLeft   []*Promo            `json:"promo_left"`
	PromoRight1 []*Promo            `json:"promo_right1"`
	PromoRight2 []*Promo            `json:"promo_right2"`
	BySection   *WeekPromoBySection `json:"by_section"`
}

// GetWeekPromoParams definisce i parametri query opzionali per GetWeekPromo.
type GetWeekPromoParams struct {
	Year int `json:"year,omitempty" query:"year"`
}

// GetWeekPromoQueryParams definisce i parametri di query per GetCurrentWeekPromo.
type GetWeekPromoQueryParams struct {
	Week int `json:"week,omitempty" query:"week"`
	Year int `json:"year,omitempty" query:"year"`
}

// getSectionOrder ritorna la priorità numerica della sezione basata sull'ordine delle costanti in promo.go.
func getSectionOrder(pos PromoPosition, section string) int {
	switch pos {
	case PromoLeft:
		switch PromoLeftSections(section) {
		case PromoLeftWeek:
			return 1
		case PromoLeftRaccomended:
			return 2
		case PromoLeftDiscover:
			return 3
		}
	case PromoRight1:
		switch PromoRight1Sections(section) {
		case PromoRight1Show:
			return 1
		case PromoRight1Service:
			return 2
		}
	case PromoRight2:
		switch PromoRight2Sections(section) {
		case PromoRight2Top:
			return 1
		case PromoRight2Selected:
			return 2
		}
	}
	return 99
}

// isPromoActiveInWeek verifica se una promozione è valida per la settimana e l'anno specificati.
func isPromoActiveInWeek(p *Promo, targetYear int, targetWeek int) bool {
	if targetYear > 0 && p.Year > 0 && p.Year != targetYear {
		return false
	}
	// Se l'array settimane è vuoto, la promozione è attiva per tutto l'anno
	if len(p.Weeks) == 0 {
		return true
	}
	for _, w := range p.Weeks {
		if w == targetWeek {
			return true
		}
	}
	return false
}

// buildWeekPromoResponse costruisce la risposta con i tre tipi di promo ordinati per sezione.
func buildWeekPromoResponse(ctx context.Context, week int, year int) (*WeekPromoResponse, error) {
	if week < 1 || week > 53 {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "settimana non valida: deve essere compresa tra 1 e 53",
		}
	}
	if year <= 0 {
		nowYear, _ := time.Now().ISOWeek()
		year = nowYear
	}

	allPromos, err := queryPromos(ctx)
	if err != nil {
		return nil, err
	}

	resp := &WeekPromoResponse{
		Week:        week,
		Year:        year,
		PromoLeft:   make([]*Promo, 0),
		PromoRight1: make([]*Promo, 0),
		PromoRight2: make([]*Promo, 0),
		BySection: &WeekPromoBySection{
			PromoLeft: PromoLeftGroup{
				Week:        make([]*Promo, 0),
				Raccomended: make([]*Promo, 0),
				Discover:    make([]*Promo, 0),
			},
			PromoRight1: PromoRight1Group{
				Show:    make([]*Promo, 0),
				Service: make([]*Promo, 0),
			},
			PromoRight2: PromoRight2Group{
				Top:      make([]*Promo, 0),
				Selected: make([]*Promo, 0),
			},
		},
	}

	for _, p := range allPromos {
		if !isPromoActiveInWeek(p, year, week) {
			continue
		}
		switch p.Position {
		case PromoLeft:
			resp.PromoLeft = append(resp.PromoLeft, p)
			switch PromoLeftSections(p.Section) {
			case PromoLeftWeek:
				resp.BySection.PromoLeft.Week = append(resp.BySection.PromoLeft.Week, p)
			case PromoLeftRaccomended:
				resp.BySection.PromoLeft.Raccomended = append(resp.BySection.PromoLeft.Raccomended, p)
			case PromoLeftDiscover:
				resp.BySection.PromoLeft.Discover = append(resp.BySection.PromoLeft.Discover, p)
			}
		case PromoRight1:
			resp.PromoRight1 = append(resp.PromoRight1, p)
			switch PromoRight1Sections(p.Section) {
			case PromoRight1Show:
				resp.BySection.PromoRight1.Show = append(resp.BySection.PromoRight1.Show, p)
			case PromoRight1Service:
				resp.BySection.PromoRight1.Service = append(resp.BySection.PromoRight1.Service, p)
			}
		case PromoRight2:
			resp.PromoRight2 = append(resp.PromoRight2, p)
			switch PromoRight2Sections(p.Section) {
			case PromoRight2Top:
				resp.BySection.PromoRight2.Top = append(resp.BySection.PromoRight2.Top, p)
			case PromoRight2Selected:
				resp.BySection.PromoRight2.Selected = append(resp.BySection.PromoRight2.Selected, p)
			}
		}
	}

	sortPromosBySection := func(promos []*Promo, pos PromoPosition) {
		sort.SliceStable(promos, func(i, j int) bool {
			ordI := getSectionOrder(pos, promos[i].Section)
			ordJ := getSectionOrder(pos, promos[j].Section)
			if ordI != ordJ {
				return ordI < ordJ
			}
			return promos[i].ID < promos[j].ID
		})
	}

	sortPromosBySection(resp.PromoLeft, PromoLeft)
	sortPromosBySection(resp.PromoRight1, PromoRight1)
	sortPromosBySection(resp.PromoRight2, PromoRight2)

	return resp, nil
}

// GetWeekPromo ritorna un JSON con i tre tipi di promo (PromoLeft, PromoRight1, PromoRight2)
// filtrati per la settimana e l'anno specificati, ordinati per le sezioni dichiarate in promo.go.
//
//encore:api public method=GET path=/promo/week/:week
func GetWeekPromo(ctx context.Context, week int, params *GetWeekPromoParams) (*WeekPromoResponse, error) {
	year := 0
	if params != nil {
		year = params.Year
	}
	return buildWeekPromoResponse(ctx, week, year)
}

// GetCurrentWeekPromo ritorna le promozioni della settimana (con settimana e anno opzionali via query ?week=W&year=Y).
//
//encore:api public method=GET path=/promo/week
func GetCurrentWeekPromo(ctx context.Context, params *GetWeekPromoQueryParams) (*WeekPromoResponse, error) {
	nowYear, nowWeek := time.Now().ISOWeek()
	week := nowWeek
	year := nowYear
	if params != nil {
		if params.Week > 0 {
			week = params.Week
		}
		if params.Year > 0 {
			year = params.Year
		}
	}
	return buildWeekPromoResponse(ctx, week, year)
}

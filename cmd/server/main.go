package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/zeshi09/dialectApi/ent"
	"github.com/zeshi09/dialectApi/ent/location"
)

var (
	templates *template.Template
	dbClient  *ent.Client
)

func main() {
	ctx := context.Background()

	var err error
	templates, err = template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatalf("templates error: %v", err)
	}

	dbClient = connectDB(ctx)
	defer dbClient.Close()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/card", handleCard)
	http.HandleFunc("/detail", handleDetail)
	http.HandleFunc("/close", handleClose)
	http.HandleFunc("/api/locations", handleLocations)
	http.HandleFunc("/api/filters", handleFilters)
	http.HandleFunc("/api/markers", handleMarkers)

	// Статические файлы (изображения)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("starting on 0.0.0.0:9090")

	if err := http.ListenAndServe("0.0.0.0:9090", nil); err != nil {
		log.Fatalf("starting server error: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}
}

func handleCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	location, err := dbClient.Location.Get(ctx, id)
	if err != nil {
		log.Printf("error fetching location %d: %v", id, err)
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`
		<div class="card"
		     hx-get="/detail?id=%d"
		     hx-target="#card-detail"
		     hx-swap="innerHTML">
			<h3>%s</h3>
			<p><strong>Определение:</strong> %s</p>
			<p><small>%s, %s</small></p>
		</div>
	`, location.ID, location.Chrononym, location.Definition, location.District, location.Selsovet)

	w.Write([]byte(html))
}

func handleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	location, err := dbClient.Location.Get(ctx, id)
	if err != nil {
		log.Printf("error fetching location %d: %v", id, err)
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Формируем HTML, показывая только непустые поля
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString(`<div class="card-detail-overlay" id="detail-overlay">`)
	htmlBuilder.WriteString(`<div class="corner-bottom-left"></div>`)
	htmlBuilder.WriteString(`<div class="corner-bottom-right"></div>`)
	htmlBuilder.WriteString(`<div class="detail-controls">`)
	htmlBuilder.WriteString(`<button class="nav-btn" onclick="navigatePrev()" title="Предыдущее">◀</button>`)
	htmlBuilder.WriteString(`<button class="nav-btn" onclick="navigateNext()" title="Следующее">▶</button>`)
	htmlBuilder.WriteString(`<button class="copy-btn" onclick="copyDetailText()" title="Копировать текст"></button>`)
	htmlBuilder.WriteString(`<button class="close-btn" onclick="closeCardSmoothly()">✖ Закрыть</button>`)
	htmlBuilder.WriteString(`</div>`)
	htmlBuilder.WriteString(fmt.Sprintf(`<h2>%s</h2>`, location.Chrononym))

	// Определение - всегда показываем
	htmlBuilder.WriteString(fmt.Sprintf(`<p><strong>Определение:</strong> %s</p>`, location.Definition))

	// Контекст - только если не пустой
	if location.Context != "" {
		htmlBuilder.WriteString(fmt.Sprintf(`<p><strong>Контекст:</strong> %s</p>`, location.Context))
	}

	// Район - всегда показываем
	htmlBuilder.WriteString(fmt.Sprintf(`<p><strong>Район:</strong> %s</p>`, location.District))

	// Сельсовет - всегда показываем
	htmlBuilder.WriteString(fmt.Sprintf(`<p><strong>Сельсовет:</strong> %s</p>`, location.Selsovet))

	// Год - только если не пустой
	if location.Year != "" {
		htmlBuilder.WriteString(fmt.Sprintf(`<p><strong>Год:</strong> %s</p>`, location.Year))
	}

	// Комментарий - только если не пустой
	if location.Comment != "" {
		htmlBuilder.WriteString(fmt.Sprintf(`<p><strong>Комментарий:</strong> %s</p>`, location.Comment))
	}

	htmlBuilder.WriteString(`</div>`)

	w.Write([]byte(htmlBuilder.String()))
}

func handleClose(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(``))
}

func handleLocations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Параметры пагинации
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 100 // По умолчанию загружаем 100 карточек за раз

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Параметры фильтрации
	searchQuery := r.URL.Query().Get("search")
	districtFilter := r.URL.Query().Get("district")
	selsovetFilter := r.URL.Query().Get("selsovet")

	// Строим запрос с фильтрами
	query := dbClient.Location.Query()

	if searchQuery != "" {
		query = query.Where(location.ChrononymContainsFold(searchQuery))
	}
	if districtFilter != "" {
		query = query.Where(location.DistrictEQ(districtFilter))
	}
	if selsovetFilter != "" {
		query = query.Where(location.SelsovetEQ(selsovetFilter))
	}

	// Получаем общее количество записей с учетом фильтров
	totalCount, err := query.Clone().Count(ctx)
	if err != nil {
		log.Printf("error counting locations: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Загружаем порцию локаций (сортировка по алфавиту)
	locations, err := query.
		Order(ent.Asc(location.FieldChrononym)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		log.Printf("error fetching locations: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Добавляем скрытый элемент с общим количеством
	w.Write([]byte(fmt.Sprintf(`<div id="total-count" data-count="%d" style="display:none;"></div>`, totalCount)))

	for _, loc := range locations {
		html := fmt.Sprintf(`
		<div class="card"
		     hx-get="/detail?id=%d"
		     hx-target="#card-detail"
		     hx-swap="innerHTML"
		     onclick="openCard(%d)">
			<h3>%s</h3>
			<p><strong>Определение:</strong> %s</p>
			<p><small>%s, %s</small></p>
		</div>
		`, loc.ID, loc.ID, loc.Chrononym, loc.Definition, loc.District, loc.Selsovet)
		w.Write([]byte(html))
	}

	// Если есть еще данные, добавляем триггер для загрузки следующей порции
	nextOffset := offset + limit
	if nextOffset < totalCount {
		// Формируем URL с учетом фильтров
		var urlBuilder strings.Builder
		urlBuilder.WriteString(fmt.Sprintf("/api/locations?offset=%d&limit=%d", nextOffset, limit))

		if searchQuery != "" {
			urlBuilder.WriteString(fmt.Sprintf("&search=%s", searchQuery))
		}
		if districtFilter != "" {
			urlBuilder.WriteString(fmt.Sprintf("&district=%s", districtFilter))
		}
		if selsovetFilter != "" {
			urlBuilder.WriteString(fmt.Sprintf("&selsovet=%s", selsovetFilter))
		}

		loadMoreHTML := fmt.Sprintf(`
		<div id="load-more"
		     hx-get="%s"
		     hx-trigger="intersect once"
		     hx-swap="outerHTML">
			<p style="text-align: center; color: #666;">Загрузка...</p>
		</div>
		`, urlBuilder.String())
		w.Write([]byte(loadMoreHTML))
	} else {
		// Все данные загружены
		endHTML := `<p style="text-align: center; color: #999; margin-top: 20px;">Все карточки загружены</p>`
		w.Write([]byte(endHTML))
	}
}

func handleFilters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем все локации и извлекаем уникальные значения
	locations, err := dbClient.Location.Query().All(ctx)
	if err != nil {
		log.Printf("error fetching locations: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Собираем уникальные районы и сельсоветы
	districtMap := make(map[string]bool)
	selsovetMap := make(map[string]bool)

	for _, loc := range locations {
		districtMap[loc.District] = true
		selsovetMap[loc.Selsovet] = true
	}

	// Конвертируем в отсортированные слайсы
	var districts []string
	for district := range districtMap {
		districts = append(districts, district)
	}
	sort.Strings(districts)

	var selsovets []string
	for selsovet := range selsovetMap {
		selsovets = append(selsovets, selsovet)
	}
	sort.Strings(selsovets)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Формируем HTML с селектами для районов
	var htmlBuilder strings.Builder
	for _, district := range districts {
		htmlBuilder.WriteString(fmt.Sprintf(`<option value="%s">%s</option>`, district, district))
	}
	districtOptions := htmlBuilder.String()

	// Формируем HTML с селектами для сельсоветов
	htmlBuilder.Reset()
	for _, selsovet := range selsovets {
		htmlBuilder.WriteString(fmt.Sprintf(`<option value="%s">%s</option>`, selsovet, selsovet))
	}
	selsovetOptions := htmlBuilder.String()

	// Отправляем JSON-подобный ответ, но в виде HTML атрибутов
	response := fmt.Sprintf(`<div id="filter-options" data-districts='%s' data-selsovets='%s'></div>`,
		districtOptions, selsovetOptions)
	w.Write([]byte(response))
}

func handleMarkers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Параметры фильтрации
	searchQuery := r.URL.Query().Get("search")
	districtFilter := r.URL.Query().Get("district")
	selsovetFilter := r.URL.Query().Get("selsovet")

	// Строим запрос с фильтрами
	query := dbClient.Location.Query()

	if searchQuery != "" {
		query = query.Where(location.ChrononymContainsFold(searchQuery))
	}
	if districtFilter != "" {
		query = query.Where(location.DistrictEQ(districtFilter))
	}
	if selsovetFilter != "" {
		query = query.Where(location.SelsovetEQ(selsovetFilter))
	}

	// Получаем все локации с учетом фильтров
	locations, err := query.All(ctx)
	if err != nil {
		log.Printf("error fetching locations for markers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Группируем по координатам
	type MarkerData struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		District  string  `json:"district"`
		Selsovet  string  `json:"selsovet"`
	}

	// Используем map для хранения уникальных координат
	markerMap := make(map[string]MarkerData)

	for _, loc := range locations {
		// Пропускаем записи без координат (проверяем на 0)
		if loc.Latitude == 0 && loc.Longitude == 0 {
			continue
		}

		// Создаем ключ из координат
		key := fmt.Sprintf("%.6f_%.6f", loc.Latitude, loc.Longitude)

		// Добавляем только если такой ключ еще не существует
		if _, exists := markerMap[key]; !exists {
			markerMap[key] = MarkerData{
				Latitude:  loc.Latitude,
				Longitude: loc.Longitude,
				District:  loc.District,
				Selsovet:  loc.Selsovet,
			}
		}
	}

	// Преобразуем map в slice
	markers := make([]MarkerData, 0, len(markerMap))
	for _, marker := range markerMap {
		markers = append(markers, marker)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(markers); err != nil {
		log.Printf("error encoding markers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

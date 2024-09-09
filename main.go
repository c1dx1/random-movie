package main

import (
    "encoding/json"
    "fmt"
    "github.com/gorilla/mux"
    "html/template"
    "io"
    "log"
    "net/http"
    "strconv"
    "sync"
)

type MovieData struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Type string `json:"type"`
	Year int `json:"year"`
	Rating struct {
		KP float32 `json:"kp"`
	} `json:"rating"`
	Genres[] struct {
		Genre string `json:"name"`
	} `json:"genres"`
	Countries[] struct {
		Country string `json:"name"`
	} `json:"countries"`
	Poster struct {
		URL string `json:"url"`
	} `json:"poster"`
	ID int `json:"id"`
	IsSeries bool `json:"isSeries"`
	URL string
}

func getMovieData(url string, ch chan <- *MovieData, wg *sync.WaitGroup) {
	defer wg.Done()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Ошибка создания запроса", err)
		ch <- nil
		return
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", "...")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Ошибка выполнения запроса", err)
		ch <- nil
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		log.Printf("Ошибка API: %s, статус код: %d, тело ответа: %s", res.Status, res.StatusCode, string(body))
		ch <- nil
		return
	}

	var movieData MovieData

	if err := json.NewDecoder(res.Body).Decode(&movieData); err != nil {
		log.Println("Ошибка декодирования JSON", err)
		ch <- nil
		return
	}

	ch <- &movieData
}

func formHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("templates/index.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl.Execute(w, nil)
}

func movieHandler (w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	url := "https://api.kinopoisk.dev/v1.4/movie/random?notNullFields=name&notNullFields=description&notNullFields=type&notNullFields=year&notNullFields=rating.kp&notNullFields=rating.imdb&notNullFields=genres.name&notNullFields=countries.name&notNullFields=poster.url"

	typeM := r.FormValue("typeM")
	yearInput, _ := strconv.Atoi(r.FormValue("year"))
    year := fmt.Sprintf("&year=%v-%v", yearInput, 2024)
	ratingKPinput, _ := strconv.ParseFloat(r.FormValue("KP"), 32)
    ratingKP := fmt.Sprintf("&rating.kp=%v-%v", ratingKPinput, 10)
	genre := r.FormValue("genre")
	country := r.FormValue("country")

	if typeM != "" {url += typeM}
	url += year
	url += ratingKP
	if genre != "" {url += genre}
	if country != "" {url += country}

	movieChan := make(chan *MovieData)
	var wg sync.WaitGroup
	wg.Add(1)
	go getMovieData(url, movieChan, &wg)

	go func(){
		wg.Wait()
		close(movieChan)
	}()

	movieData := <- movieChan
	if movieData == nil {
		http.Error(w, "Ошибка при получении данных о фильме", http.StatusInternalServerError)
		return
	}

	var typeForURL string

	if movieData.IsSeries == true {
		typeForURL = "series"
	} else {
		typeForURL = "film"
	}

	movieData.URL = fmt.Sprintf("https://kinopoisk.ru/%s/%v", typeForURL, movieData.ID)

	tmpl, err := template.ParseFiles("templates/movie.html")
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	err = tmpl.Execute(w, movieData)
	if err != nil {
		log.Printf("Ошибка при выполнении шаблона: %v", err)
    	http.Error(w, "Ошибка при выполнении шаблона", http.StatusInternalServerError)
	}
}

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	r.HandleFunc("/", formHandler)
	r.HandleFunc("/postform", movieHandler)

	log.Println("Start server on localhost")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error start server: %v", err)
	}
}
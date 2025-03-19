package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

//go:embed templates/*
var indexPage embed.FS

type PexelPhoto struct {
	Photographer     string `json:"photographer"`
	Photographer_url string `json:"photographer_url"`
	Alt              string `json:"alt"`
	Src              struct {
		Original string `json:"original"`
	} `json:"src"`
}

type PexelPhotoString struct {
	Photographer     string
	Photographer_url string
	Alt              string
	Src              string
}

type ApiResponse struct {
	Photos []PexelPhoto `json:"photos"`
}

func callApi() ([]byte, error) {
	pexApi := os.Getenv("PEXELS_KEY")
	url := "https://api.pexels.com/v1/curated?per_page=1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Couldn't create request")
		return []byte(""), err
	}

	req.Header.Add("Authorization", pexApi)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error doing the request")
		return []byte(""), err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading the response body")
		return []byte(""), err
	}

	return body, nil
}

func getPhotoJson(bodyOfApi []byte) (PexelPhoto, error) {
	var response ApiResponse
	err := json.Unmarshal([]byte(bodyOfApi), &response)
	if err != nil {
		log.Fatal("Could not unmarshal json response, ", err)
		return PexelPhoto{}, err
	}

	if len(response.Photos) > 0 {
		PhotoData := PexelPhoto{
			Photographer:     response.Photos[0].Photographer,
			Photographer_url: response.Photos[0].Photographer_url,
			Alt:              response.Photos[0].Alt,
			Src:              response.Photos[0].Src,
		}
		return PhotoData, nil
	} else {
		fmt.Println("Couldn't find photo :c")
		return PexelPhoto{}, err
	}
}

func mainView(w http.ResponseWriter, r *http.Request) {
	apiData, err := callApi()
	if err != nil {
		log.Fatal(err)
	}

	photoJsonData, err := getPhotoJson(apiData)
	if err != nil {
		log.Fatal(err)
	}

	Data := PexelPhotoString{
		photoJsonData.Photographer,
		photoJsonData.Photographer_url,
		photoJsonData.Alt,
		photoJsonData.Src.Original,
	}

	tmpl := template.Must(template.ParseFS(indexPage, "templates/index.html"))

	tmpl.Execute(w, Data)
}

func main() {
	port := os.Getenv("PORT")

	http.HandleFunc("/", mainView)

	fmt.Println("Server starting on port: ", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Couldn't start server on port :8080")
	}
}

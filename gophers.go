package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/gorilla/mux"
)

const G = "g"
const GE = "ge"
const XR = "xr"
const Q = "q"
const U = "u"
const OGO = "ogo"

type CustomError struct {
	Msg string `json:"error-msg"`
}

type InputWord struct {
	EnglishWord string `json:"english-word"`
}

type InputSentence struct {
	EnglishSentence string `json:"english-sentence"`
}

type OutputWord struct {
	GopherWord string `json:"gopher-word"`
}

type OutputSentence struct {
	GopherSentence string `json:"gopher-sentence"`
}

type History struct {
	Data map[string]string `json:"history"`
}

var translations map[string]string

var customErr CustomError

var js []byte
var err error

func translateWord(w http.ResponseWriter, req *http.Request) {

	var data InputWord
	var result OutputWord

	if err_ := json.NewDecoder(req.Body).Decode(&data); err_ != nil {
		log.Println(err_)
	}

	if data.EnglishWord == "" {
		customErr.Msg = "Input Data Error"
		js, err = json.Marshal(customErr)
	} else {
		translatedWord := translate(data.EnglishWord)

		if translations == nil {
			translations = make(map[string]string)
		}
		translations[data.EnglishWord] = translatedWord
		
		result.GopherWord = translatedWord

		js, err = json.Marshal(result)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func translateSentence(w http.ResponseWriter, req *http.Request) {

	var data InputSentence
	var translatedSentence string
	if err_ := json.NewDecoder(req.Body).Decode(&data); err_ != nil {
		log.Println(err_)
	}
	if data.EnglishSentence == "" {
		customErr.Msg = "Input Data Error"

		js, err = json.Marshal(customErr)

	} else {
		words := strings.Split(data.EnglishSentence, " ")

		for _, word := range words {
			translatedSentence += translate(word) + " "
		}

		if translations == nil {
			translations = make(map[string]string)
		}

		translations[data.EnglishSentence] = strings.TrimSpace(translatedSentence)

		var result OutputSentence
		result.GopherSentence = strings.TrimSpace(translatedSentence)

		js, err = json.Marshal(result)

	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func history(w http.ResponseWriter, req *http.Request) {

	var history History
	history.Data = translations
	js, err := json.Marshal(history)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func translate(word string) string {
	var result string
	var seqOfConsonant string
	var endIndexOfConsonant int
	r := []rune(strings.ToLower(word))
	if isVowel(r[0]) {
		result = G + word
	} else if len(word) > 1 && strings.ToLower(word)[0:2] == XR {
		result = GE + word
	} else if !isVowel(r[0]) {
		endIndexOfConsonant = countSeqOfConsonant(r)
		seqOfConsonant = word[:endIndexOfConsonant]

		if seqOfConsonant[len(seqOfConsonant)-1:] == Q && word[endIndexOfConsonant:endIndexOfConsonant+1] == U {
			seqOfConsonant = seqOfConsonant + U
			result = word[endIndexOfConsonant+1:] + seqOfConsonant + OGO
		} else {
			result = word[endIndexOfConsonant:] + seqOfConsonant + OGO
		}
	}
	return result
}

func isVowel(char rune) bool {
	var result bool
	switch char {
	case 'a', 'e', 'i', 'o', 'u', 'y':
		result = true
	}
	return result
}

func countSeqOfConsonant(word []rune) int {
	var result int = 1

	for i := 1; i < len(word); i++ {
		if !isVowel(word[i]) {
			result++
		} else {
			break
		}
	}
	return result
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/word", translateWord).Methods("POST")
	r.HandleFunc("/sentence", translateSentence).Methods("POST")
	r.HandleFunc("/history", history).Methods("GET")
	http.Handle("/", r)
	var port string = os.Args[1]

	fmt.Printf("Starting server on PORT: " + port + "\n")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

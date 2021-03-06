package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	alfred "github.com/pascalw/go-alfred"
)

type Request struct {
	URL   string
	Lang  string
	Query string
}

type Translation struct {
	Meaning string
	Href    string
	Phrase  []string
}

func main() {
	base := "http://www.linguee.fr/%s/search?qe=%s&source=auto"

	args := params(base, os.Args)

	selection := filterDocument(transformToDocument(request(args)))

	response := alfred.NewResponse()

	match := false

	for index, translation := range parseTranslations(selection) {

		if strings.ToLower(args.Query) == strings.ToLower(translation.Meaning) {
			match = true
		}

		response.AddItem(&alfred.AlfredResponseItem{
			Valid:    true,
			Uid:      strconv.Itoa(index),
			Title:    translation.Meaning,
			Subtitle: strings.Join(translation.Phrase, ", "),
			Arg:      fmt.Sprintf("http://www.linguee.fr%s", translation.Href),
		})
	}

	// Add query itself to give the ability for a direct search if there is no match
	if match == false {
		origin, _ := url.QueryUnescape(args.Query)

		response.AddItem(&alfred.AlfredResponseItem{
			Valid: true,
			Uid:   "origin",
			Title: origin,
			Arg:   fmt.Sprintf("http://www.linguee.fr/french-english/search?source=auto&query=%s", url.QueryEscape(args.Query)),
		})
	}

	response.Print()
}

func params(base string, args []string) Request {
	if len(args) == 1 {
		log.Fatal("Missing Arguments")
	}

	query := args[1]
	lang := "french-english"

	if len(args) > 2 {
		lang = args[2]
	}

	return Request{base, lang, url.QueryEscape(query)}
}

func request(req Request) *iconv.Reader {
	// Http request
	res, err := http.Get(fmt.Sprintf(req.URL, req.Lang, req.Query))
	if err != nil {
		log.Fatal(err)
	}

	// charset conversion
	body, err := iconv.NewReader(res.Body, "iso-8859-15", "utf-8")
	if err != nil {
		log.Fatal(err)
	}

	return body
}

func transformToDocument(reader *iconv.Reader) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(reader)

	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func filterDocument(doc *goquery.Document) *goquery.Selection {
	doc.Find(".wordtype, .sep, .grammar_info").ReplaceWithHtml("")
	return doc.Find(".autocompletion_item")
}

func parseTranslations(elements *goquery.Selection) (results []Translation) {
	elements.Each(func(index int, element *goquery.Selection) {
		results = append(results, Translation{parseMeaning(element), parseHref(element), parsePhrase(element)})
	})

	return
}

func parseMeaning(selection *goquery.Selection) string {
	return strings.TrimSpace(selection.Find(".main_item").Text())
}

func parseHref(selection *goquery.Selection) string {
	href, _ := selection.Find(".main_item").Attr("href")
	return href
}

func parsePhrase(selection *goquery.Selection) (result []string) {
	selection.Find(".translation_item").Each(func(index int, meaning *goquery.Selection) {
		result = append(result, strings.TrimSpace(meaning.Text()))
	})

	return
}

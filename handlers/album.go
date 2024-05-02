package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/rramiachraf/dumb/data"
	"github.com/rramiachraf/dumb/utils"
	"github.com/rramiachraf/dumb/views"
)

func album(l *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		artist := mux.Vars(r)["artist"]
		albumName := mux.Vars(r)["albumName"]

		id := fmt.Sprintf("%s/%s", artist, albumName)

		if a, err := getCache[data.Album](id); err == nil {
			views.AlbumPage(a).Render(context.Background(), w)
			return
		}

		url := fmt.Sprintf("https://genius.com/albums/%s/%s", artist, albumName)

		resp, err := sendRequest(url)
		if err != nil {
			l.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			views.ErrorPage(500, "cannot reach Genius servers").Render(context.Background(), w)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			w.WriteHeader(http.StatusNotFound)
			views.ErrorPage(404, "page not found").Render(context.Background(), w)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			l.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			views.ErrorPage(500, "something went wrong").Render(context.Background(), w)
			return
		}

		cf := doc.Find(".cloudflare_content").Length()
		if cf > 0 {
			l.Error("cloudflare got in the way")
			views.ErrorPage(500, "i'll fix this later #21").Render(context.Background(), w)
			return
		}

		var a data.Album
		if err = a.Parse(doc); err != nil {
			l.Error(err.Error())
		}

		views.AlbumPage(a).Render(context.Background(), w)

		if err = setCache(id, a); err != nil {
			l.Error(err.Error())
		}
	}
}

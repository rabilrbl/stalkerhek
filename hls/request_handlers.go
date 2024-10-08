package hls

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Handles '/iptv' requests
func playlistHandler(w http.ResponseWriter, r *http.Request) {
	// Content-Type for m3u playlist
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	// Set file name
	w.Header().Set("Content-Disposition", "attachment; filename=\"playlist_"+r.Host+".m3u\"")
	w.WriteHeader(http.StatusOK)
	
	var schema string
	if r.TLS == nil {
		schema = "http"
	} else {
		schema = "https"
	}
	host_url := schema + "://" + r.Host

	fmt.Fprintln(w, "#EXTM3U")
	for _, title := range sortedChannels {
		link := host_url + "/iptv/" + url.PathEscape(title)
		logo := host_url + "/logo/" + url.PathEscape(title)

		fmt.Fprintf(w, "#EXTINF:-1 tvg-logo=\"%s\" group-title=\"%s\", %s\n%s\n", logo, playlist[title].Genre, title, link)
	}
}

// Handles '/iptv/' requests
func channelHandler(w http.ResponseWriter, r *http.Request) {
	cr, err := getContentRequest(w, r, "/iptv/")
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Lock channel's mux
	cr.ChannelRef.Mux.Lock()

	// Keep track on channel access time
	if err = cr.ChannelRef.validate(); err != nil {
		cr.ChannelRef.Mux.Unlock()
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if os.Getenv("REDIRECT_CHANNEL_LINKS") == "true" {
		cr.ChannelRef.Mux.Unlock()
		http.Redirect(w, r, cr.ChannelRef.Link, http.StatusFound)
		return
	}

	// Handle content
	handleContent(cr)
}

// Handles '/logo/' requests
func logoHandler(w http.ResponseWriter, r *http.Request) {
	cr, err := getContentRequest(w, r, "/logo/")
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Lock
	cr.ChannelRef.Logo.Mux.Lock()

	// Retrieve from Stalker middleware if no cache is present
	if len(cr.ChannelRef.Logo.Cache) == 0 {
		logo_link := strings.Split(cr.ChannelRef.Logo.Link, "http")
		if len(logo_link) > 2 {
			cr.ChannelRef.Logo.Link = "http" + logo_link[2]
		}
		img, contentType, err := download(cr.ChannelRef.Logo.Link)
		if err != nil {
			cr.ChannelRef.Logo.Mux.Unlock()
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		cr.ChannelRef.Logo.Cache = img
		cr.ChannelRef.Logo.CacheContentType = contentType
	}

	// Create local copy so we don't need thread syncrhonization
	logo := *cr.ChannelRef.Logo

	// Unlock
	cr.ChannelRef.Logo.Mux.Unlock()

	w.Header().Set("Content-Type", logo.CacheContentType)
	w.Write(logo.Cache)
}

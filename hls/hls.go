package hls

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/rabilrbl/stalkerhek/stalker"
)

var playlist map[string]*Channel
var sortedChannels []string

// Start starts main routine.
func Start(chs map[string]*stalker.Channel, bind string) {
	// Initialize playlist
	playlist = make(map[string]*Channel)
	sortedChannels = make([]string, 0, len(chs))
	for k, v := range chs {
		playlist[k] = &Channel{
			StalkerChannel: v,
			Mux:            &sync.Mutex{},
			Logo: &Logo{
				Mux:  &sync.Mutex{},
				Link: v.Logo(),
			},
			Genre: v.Genre(),
		}
		sortedChannels = append(sortedChannels, k)
	}
	sort.Strings(sortedChannels)

	mux := http.NewServeMux()
	mux.HandleFunc("/iptv", playlistHandler)
	mux.HandleFunc("/iptv/", channelHandler)
	mux.HandleFunc("/logo/", logoHandler)

	log.Println("HLS service should be started!")
	panic(http.ListenAndServe(bind, mux))
}

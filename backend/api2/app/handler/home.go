package handler

import (
	"main/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Home(ctx *gin.Context) {
	torrentSearch, torrentSearchExist := ctx.GetQuery("s")

	if !torrentSearchExist {
		ctx.HTML(http.StatusOK, "home/index.html", nil)
		return
	}

	htmxReqHead := ctx.GetHeader("HX-Request")

	var torrentList *utils.TorrentListStruct
	var err error

	torrentList, err = utils.GetTorrentList(torrentSearch)

	if err != nil {
		ctx.HTML(http.StatusOK, "home/torrentList.html", gin.H{
			"message": "Could not fetch torrents: " + err.Error(),
		})
		return
	}

	if htmxReqHead == "true" {
		ctx.HTML(http.StatusOK, "home/torrentList.html", gin.H{"torrentList": torrentList})
		return
	}

	ctx.HTML(http.StatusOK, "home/index.html", gin.H{"s": true, "input": torrentSearch, "torrentList": torrentList})
}

package controller

import (
	"363project/controller/service"
	"363project/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func USSDHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	userIdCtx := r.Context().Value("id")
	idUint, _ := strconv.ParseUint(fmt.Sprint(userIdCtx), 10, 32)

	for {
		var req struct {
			Option int `json:"option"`
		}

		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		var res struct {
			Description string   `json:"description"`
			Menu        []string `json:"menu"`
			End         bool     `json:"end"`
		}

		switch req.Option {
		case 1: // Hot Promo
			list, _ := service.ShowPenawaran("promo")
			res.Description = "Hot Promo Spesial:"
			res.Menu = formatMenu(list)
		case 2: // Internet Harian
			list, _ := service.ShowPenawaran("harian")
			res.Description = "Paket Internet Harian:"
			res.Menu = formatMenu(list)
		case 3: // Internet Mingguan
			list, _ := service.ShowPenawaran("mingguan")
			res.Description = "Paket Internet Mingguan:"
			res.Menu = formatMenu(list)
		case 4: // Internet Bulanan
			list, _ := service.ShowPenawaran("bulanan")
			res.Description = "Paket Internet Bulanan:"
			res.Menu = formatMenu(list)
		case 5: // Combo Internet + Telpon
			list, _ := service.ShowPenawaran("combo")
			res.Description = "Paket Combo Seru:"
			res.Menu = formatMenu(list)
		case 6: // Paket Malam
			list, _ := service.ShowPenawaran("malam")
			res.Description = "Paket Internet Malam:"
			res.Menu = formatMenu(list)
		case 7: // Paket Game & Streaming
			list, _ := service.ShowPenawaran("game")
			res.Description = "Paket Game & Streaming:"
			res.Menu = formatMenu(list)
		case 8: // Cek Pulsa
			pulsa, _ := service.CheckPulsa(uint(idUint))
			res.Description = fmt.Sprintf("Sisa Pulsa Anda: Rp%.2f", pulsa)
			res.End = true
		case 9: // Cek Kuota
			kuota, _ := service.CheckKuota(uint(idUint))
			res.Description = fmt.Sprintf("Sisa Kuota Anda: %.2f GB", kuota/1000)
			res.End = true
		default:
			res.Description = "Pilihan tidak valid."
			res.End = true
		}

		conn.WriteJSON(res)
		if res.End {
			break
		}
	}
}

func formatMenu(penawaran []model.Penawaran) []string {
	var m []string
	for _, p := range penawaran {
		m = append(m, fmt.Sprintf("%dGB/%dHr/Rp%d", p.Jumlah/1000, p.Durasi, p.Harga))
	}
	return m
}

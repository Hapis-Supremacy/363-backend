package controller

import (
	"363project/controller/service"
	"363project/model"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type USSDResponse struct {
	Description string   `json:"description"`
	Menu        []string `json:"menu"`
	End         bool     `json:"end"`
}

func CloseAndReset(conn *websocket.Conn, message string) {
	res := USSDResponse{
		Description: message,
		Menu:        []string{},
		End:         true,
	}
	conn.WriteJSON(res)
	conn.Close()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func USSDHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	ussd, ok := r.Context().Value("ussd").(model.USSDCookie)
	if !ok {
		CloseAndReset(conn, "Sesi tidak valid.")
		return
	}

	// State internal
	ussd.Step = 0
	var currentOffers []model.Penawaran

	// --- INISIALISASI: Kirim Menu Utama saat pertama kali buka ---
	initialMenu := USSDResponse{
		Description: "Layanan USSD *858#",
		Menu: []string{
			"1.Hot Promo", "2.Internet Harian", "3.Internet Mingguan",
			"4.Internet Bulanan", "5.Combo Internet + Telpon", "6.Paket Malam",
			"7.Paket Game & Streaming", "8.Cek Pulsa", "9.Cek Kuota",
		},
		End: false,
	}
	conn.WriteJSON(initialMenu)

	for {
		var req struct {
			Option int `json:"option"`
		}

		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		if ussd.Step == 0 {
			// --- LOGIKA STEP 0 (Sama dengan kode kamu) ---
			switch req.Option {
			case 1, 2, 3, 4, 5, 6, 7:
				categories := map[int]string{
					1: "promo", 2: "harian", 3: "mingguan",
					4: "bulanan", 5: "combo", 6: "malam", 7: "game",
				}

				list, err := service.ShowPenawaran(categories[req.Option])
				if err != nil {
					CloseAndReset(conn, "Maaf, paket tidak tersedia saat ini.")
					return
				}

				currentOffers = list
				ussd.Step = 1
				data := ussd
				updateUSSDCookie(w, r, data)

				conn.WriteJSON(USSDResponse{
					Description: "Pilih paket yang ingin dibeli:",
					Menu:        formatMenu(list),
					End:         false,
				})

			case 8: // Cek Pulsa
				pulsa, _ := service.CheckPulsa(ussd.UserId)
				CloseAndReset(conn, fmt.Sprintf("Sisa Pulsa Anda: Rp%.2f", pulsa))
				return

			case 9: // Cek Kuota
				kuota, _ := service.CheckKuota(ussd.UserId)
				CloseAndReset(conn, fmt.Sprintf("Sisa Kuota Anda: %.2f GB", kuota/1000000000))
				return

			default:
				CloseAndReset(conn, "Pilihan tidak valid.")
				return
			}

		} else if ussd.Step == 1 {
			// --- LOGIKA STEP 1 (Sama dengan kode kamu) ---
			index := req.Option - 1
			if index < 0 || index >= len(currentOffers) {
				CloseAndReset(conn, "Pilihan paket tidak valid.")
				return
			}

			selectedPackage := currentOffers[index]
			_, err := service.BuyPackage(selectedPackage, ussd.UserId)

			if err != nil {
				CloseAndReset(conn, "Gagal: "+err.Error())
			} else {
				CloseAndReset(conn, fmt.Sprintf("Terima kasih! Paket %s Anda sudah aktif.\nSelamat menikmati!", selectedPackage.Jenis))
			}
			return
		}
	}
}

func updateUSSDCookie(w http.ResponseWriter, r *http.Request, USSD model.USSDCookie) {
	jsonBytes, _ := json.Marshal(USSD)
	cookie := &http.Cookie{
		Name:     "ussd_state",
		Value:    string(jsonBytes),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24,
	}
	http.SetCookie(w, cookie)
}

func formatMenu(penawaran []model.Penawaran) []string {
	var m []string
	for _, p := range penawaran {
		// Asumsi p.Jumlah dalam Byte, kita ubah ke GB
		gb := p.Jumlah / 1000000000
		m = append(m, fmt.Sprintf("%.1fGB/%dHr/Rp%.0f", gb, p.Durasi, p.Harga))
	}
	return m
}

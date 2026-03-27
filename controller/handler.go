package controller

import (
	"363project/controller/service"
	"363project/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// USSDResponse is the JSON response sent to the client over WebSocket.
type USSDResponse struct {
	Description string   `json:"description"`
	Menu        []string `json:"menu"`
	End         bool     `json:"end"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var mainMenu = USSDResponse{
	Description: "Layanan USSD *858#",
	Menu: []string{
		"Hot Promo", "Internet Harian", "Internet Mingguan",
		"Internet Bulanan", "Combo Internet + Telpon", "Paket Malam",
		"Paket Game & Streaming", "Cek Pulsa", "Cek Kuota",
	},
	End: false,
}

var categoryMap = map[int]string{
	1: "Hot Promo",
	2: "Internet Harian",
	3: "Internet Mingguan",
	4: "Internet Bulanan",
	5: "Combo Internet + Telpon",
	6: "Paket Malam",
	7: "Paket Game & Streaming",
}

// closeAndReset sends a terminal message, resets session step, and closes the connection.
func closeAndReset(conn *websocket.Conn, ussd model.USSDCookie, message string) {
	res := USSDResponse{
		Description: message,
		Menu:        []string{},
		End:         true,
	}
	if err := conn.WriteJSON(res); err != nil {
		log.Printf("closeAndReset: failed to write JSON: %v", err)
	}
	conn.Close()
}

func USSDHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("USSDHandler: websocket upgrade failed: %v", err)
			return
		}

		ussd, ok := r.Context().Value("ussd").(model.USSDCookie)
		if !ok {
			closeAndReset(conn, ussd, "Sesi tidak valid.")
			return
		}

		// currentOffers holds the package list selected in Step 0,
		// used for validation in Step 1 within the same connection.
		var currentOffers []model.Penawarans

		if err := conn.WriteJSON(mainMenu); err != nil {
			log.Printf("USSDHandler: failed to send initial menu: %v", err)
			return
		}

		for {
			var req struct {
				Option int `json:"option"`
			}
			if err := conn.ReadJSON(&req); err != nil {
				// Client disconnected or sent invalid data — exit silently.
				break
			}

			switch ussd.Step {
			case 0:
				if err := handleStep0(w, conn, &ussd, req.Option, &currentOffers); err != nil {
					return
				}
			case 1:
				handleStep1(conn, ussd, req.Option, currentOffers)
				return
			}
		}
	})
}

// handleStep0 processes the main menu selection.
// Returns an error if the connection should be terminated.
func handleStep0(
	w http.ResponseWriter,
	conn *websocket.Conn,
	ussd *model.USSDCookie,
	option int,
	currentOffers *[]model.Penawarans,
) error {
	switch {
	case option >= 1 && option <= 7:
		category := categoryMap[option]
		list, err := service.ShowPenawaran(category)
		if err != nil {
			log.Printf("handleStep0: ShowPenawaran(%s) error: %v", category, err)
			closeAndReset(conn, *ussd, "Maaf, paket tidak tersedia saat ini.")
			return fmt.Errorf("show penawaran: %w", err)
		}

		*currentOffers = list
		ussd.Step = 1
		updateUSSDCookie(w, *ussd)

		if err := conn.WriteJSON(USSDResponse{
			Description: "Pilih paket yang ingin dibeli:",
			Menu:        formatMenu(list),
			End:         false,
		}); err != nil {
			return fmt.Errorf("write package menu: %w", err)
		}

	case option == 8: // Cek Pulsa
		pulsa, err := service.CheckPulsa(ussd.UserId)
		if err != nil {
			log.Printf("handleStep0: CheckPulsa error: %v", err)
			closeAndReset(conn, *ussd, "Gagal mengambil data pulsa.")
			return fmt.Errorf("check pulsa: %w", err)
		}
		closeAndReset(conn, *ussd, fmt.Sprintf("Sisa Pulsa Anda: Rp%.2f", pulsa))
		return fmt.Errorf("done") // signal caller to return

	case option == 9: // Cek Kuota
		kuota, err := service.CheckKuota(ussd.UserId)
		if err != nil {
			log.Printf("handleStep0: CheckKuota error: %v", err)
			closeAndReset(conn, *ussd, "Gagal mengambil data kuota.")
			return fmt.Errorf("check kuota: %w", err)
		}
		closeAndReset(conn, *ussd, fmt.Sprintf("Sisa Kuota Anda: %.2f GB", float64(kuota)/1_000_000_000))
		return fmt.Errorf("done")

	default:
		closeAndReset(conn, *ussd, "Pilihan tidak valid.")
		return fmt.Errorf("invalid option: %d", option)
	}

	return nil
}

// handleStep1 processes the package selection and purchase.
func handleStep1(
	conn *websocket.Conn,
	ussd model.USSDCookie,
	option int,
	currentOffers []model.Penawarans,
) {
	index := option - 1
	if index < 0 || index >= len(currentOffers) {
		closeAndReset(conn, ussd, "Pilihan paket tidak valid.")
		return
	}

	selected := currentOffers[index]
	if _, err := service.BuyPackage(selected, ussd.UserId); err != nil {
		log.Printf("handleStep1: BuyPackage error: %v", err)
		closeAndReset(conn, ussd, "Gagal: "+err.Error())
		return
	}

	closeAndReset(conn, ussd, fmt.Sprintf(
		"Terima kasih! Paket %s Anda sudah aktif.\nSelamat menikmati!", selected.Jenis,
	))
}

// updateUSSDCookie persists the USSD session state to a cookie.
func updateUSSDCookie(w http.ResponseWriter, ussd model.USSDCookie) {
	jsonBytes, err := json.Marshal(ussd)
	if err != nil {
		log.Printf("updateUSSDCookie: marshal error: %v", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "ussd_state",
		Value:    string(jsonBytes),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24,
	})
}

// formatMenu converts a list of offers into human-readable menu strings.
func formatMenu(penawaran []model.Penawarans) []string {
	menu := make([]string, 0, len(penawaran))
	for _, p := range penawaran {
		gb := p.Jumlah / 1_000_000_000
		menu = append(menu, fmt.Sprintf("%dGB/%dHr/Rp%.0f", gb, p.Durasi, p.Harga))
	}
	return menu
}

package model

import "time"

type USSDCookie struct {
	UserId uint `json:"userId"`
	Step   int  `json:"step"`
}

type User struct {
	Id    uint
	Pulsa float64
	Paket []Paket
}

type Paket struct {
	Id     uint
	UserId uint
	Jumlah float64
	Durasi time.Time
	Jenis  string
}

type Penawaran struct {
	Id     uint
	Jumlah float64
	Durasi int
	Jenis  string
	Harga  float64
}

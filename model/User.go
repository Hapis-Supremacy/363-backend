package model

import "time"

type USSDCookie struct {
	UserId uint `json:"userId"`
	Step   int  `json:"step"`
}

type Customers struct {
	User_id      uint     `gorm:"column:user_id;primaryKey;autoIncrement"`
	Jumlah_pulsa float64  `gorm:"column:jumlah_pulsa"`
	Paket        []Kuotas `gorm:"foreignKey:user_id"`
}

type Kuotas struct {
	Id_kuota int       `gorm:"column:id_kuota;primaryKey;autoIncrement"`
	User_id  int       `gorm:"column:user_id;not null;"`
	Jumlah   int       `gorm:"column:jumlah;not null"`
	Durasi   time.Time `gorm:"column:durasi;not null"`
	Jenis    string    `gorm:"column:jenis;not null"`
}

type Penawarans struct {
	Id_penawaran int     `gorm:"column:id_penawaran;primaryKey;autoIncrement"`
	Jumlah       int     `gorm:"column:jumlah;not null"`
	Durasi       int     `gorm:"column:durasi;not null"`
	Jenis        string  `gorm:"column:jenis;not null"`
	Harga        float64 `gorm:"column:harga;not null"`
}

package service

import (
	"363project/initializer"
	"363project/model"
	"errors"
	"time"
)

func CreateAnonymousUser() (model.User, error) {
	user := model.User{
		Pulsa: 200000,
	}
	err := initializer.DB.Create(&user).Error
	if err != nil {
		return user, err
	}
	return user, nil
}

func BuyPackage(req model.Penawaran, userID uint) (model.Paket, error) {

	paket := model.Paket{
		Jumlah: req.Jumlah,
		UserId: userID,
		Durasi: time.Now().Add(time.Duration(req.Durasi) * time.Hour),
		Jenis:  req.Jenis,
	}

	tx := initializer.DB.Begin()
	var user model.User
	err := initializer.DB.First(&user, "id = ?", userID).Error
	if err != nil {
		return paket, err
	}

	if user.Pulsa < req.Harga {
		return paket, errors.New("pulsa tidak mencukupi")
	}
	user.Pulsa = user.Pulsa - req.Harga
	err = tx.Save(&user).Error
	if err != nil {
		return paket, err
	}

	err = tx.Create(&paket).Error
	if err != nil {
		return paket, err
	}
	tx.Commit()

	return paket, nil
}

func ShowPenawaran(jenis string) ([]model.Penawaran, error) {
	var Penawaran []model.Penawaran
	err := initializer.DB.Where("jenis = ?", jenis).Find(&Penawaran).Error
	if err != nil {
		return Penawaran, err
	}

	if len(Penawaran) == 0 {
		return Penawaran, errors.New("tidak ada data")
	}
	return Penawaran, nil
}

func CheckKuota(user uint) (float64, error) {
	var paket []model.Paket
	var total float64 = 0
	err := initializer.DB.Where("id = ?", user).Find(&paket).Error
	if err != nil {
		return 0, err
	}
	for _, val := range paket {
		total += val.Jumlah
	}
	return total, nil
}

func CheckPulsa(user uint) (float64, error) {
	var Pulsa float64
	err := initializer.DB.Model(&model.User{}).Select("pulsa").Where("id = ?", user).Scan(&Pulsa).Error
	if err != nil {
		return 0, err
	}
	return Pulsa, nil
}

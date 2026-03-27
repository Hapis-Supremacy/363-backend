package service

import (
	"363project/initializer"
	"363project/model"
	"errors"
	"log"
	"time"
)

var (
	ErrInsufficientBalance = errors.New("pulsa tidak mencukupi")
	ErrNoPackageFound      = errors.New("tidak ada paket tersedia")
	ErrUserNotFound        = errors.New("user tidak ditemukan")
)

// CreateAnonymousUser creates a new guest customer with a default balance.
func CreateAnonymousUser() (model.Customers, error) {
	user := model.Customers{
		Jumlah_pulsa: 200000,
	}
	if err := initializer.DB.Create(&user).Error; err != nil {
		return model.Customers{}, err
	}
	return user, nil
}

// BuyPackage deducts balance and activates a data package for the user.
// The operation is wrapped in a transaction — either both succeed or neither does.
func BuyPackage(req model.Penawarans, userID uint) (model.Kuotas, error) {
	paket := model.Kuotas{
		Jumlah:             req.Jumlah,
		User_id:            int(userID),
		Tanggal_kadaluarsa: time.Now().Add(time.Duration(req.Durasi) * time.Hour),
		Jenis:              req.Jenis,
	}

	tx := initializer.DB.Begin()
	if tx.Error != nil {
		return model.Kuotas{}, tx.Error
	}

	// Rollback helper — no-op if already committed.
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("BuyPackage: recovered from panic: %v", r)
		}
	}()

	var user model.Customers
	if err := tx.First(&user, "user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		return model.Kuotas{}, ErrUserNotFound
	}

	if user.Jumlah_pulsa < req.Harga {
		tx.Rollback()
		return model.Kuotas{}, ErrInsufficientBalance
	}

	user.Jumlah_pulsa -= req.Harga
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return model.Kuotas{}, err
	}

	if err := tx.Create(&paket).Error; err != nil {
		tx.Rollback()
		return model.Kuotas{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return model.Kuotas{}, err
	}

	return paket, nil
}

// ShowPenawaran returns up to 10 available packages for the given category.
func ShowPenawaran(jenis string) ([]model.Penawarans, error) {
	var penawaran []model.Penawarans
	if err := initializer.DB.Where("jenis = ?", jenis).Limit(10).Find(&penawaran).Error; err != nil {
		return nil, err
	}
	if len(penawaran) == 0 {
		return nil, ErrNoPackageFound
	}
	return penawaran, nil
}

// CheckKuota returns the total active quota (in bytes) for the given user,
// excluding expired packages.
func CheckKuota(userID uint) (int, error) {
	var total int
	err := initializer.DB.
		Model(&model.Kuotas{}).
		Select("COALESCE(SUM(jumlah), 0)").
		Where("user_id = ? AND tanggal_kadaluarsa > ?", userID, time.Now()).
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// CheckPulsa returns the current balance for the given user.
func CheckPulsa(userID uint) (float64, error) {
	var pulsa float64
	err := initializer.DB.
		Model(&model.Customers{}).
		Select("jumlah_pulsa").
		Where("user_id = ?", userID).
		Scan(&pulsa).Error
	if err != nil {
		return 0, err
	}
	return pulsa, nil
}

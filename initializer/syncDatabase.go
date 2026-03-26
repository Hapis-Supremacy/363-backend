package initializer

import "363project/model"

func SyncDatabase() {
	DB.AutoMigrate(&model.User{}, &model.Penawaran{}, &model.Paket{})
}

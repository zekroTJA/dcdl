package config

import "github.com/zekrotja/dcdl/models"

type ConfigProvider interface {
	Load() (err error)
	Instance() *models.ConfigModel
}

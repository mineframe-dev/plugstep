package plugins

import (
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
)

const cacheName = "plugins"

func InitCache() {
	utils.InitCache(cacheName)
}

func GetCache() *utils.Cache {
	return utils.GetCache(cacheName)
}

func FlushCache() error {
	return utils.FlushCache(cacheName)
}

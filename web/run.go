package web

import (
	"embed"
)

var (
	//go:embed html/*
	HtmlFiles embed.FS

	//go:embed static/*
	StaticFiles embed.FS
)


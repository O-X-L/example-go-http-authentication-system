package server

import "embed"

//go:embed files/build/*
//go:embed files/static/*
var EmbeddedStaticFiles embed.FS

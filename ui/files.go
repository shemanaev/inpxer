package ui

import "embed"

//go:embed static
var StaticFiles embed.FS

//go:embed templates/*
var Templates embed.FS

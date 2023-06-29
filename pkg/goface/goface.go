package goface

import (
    "log"
    "path/filepath"

    "github.com/Kagami/go-face"
)

const (
    MouseDetected = "1"
    EyeDetected = "2"
)

const (
    MouseOpen = 1
    MouseClose = 2
)

const (
    EyeOpen = 1
    EyeClose = 2
)

type Goface struct {
    Rec *face.Recognizer
}

func NewGoface(dataDir string) *Goface {
    rec, err := face.NewRecognizer(filepath.Join(dataDir, "models"))
    if err != nil {
        log.Panicf("Can't init face recognizer: %v", err)
    }
    return &Goface{Rec: rec}
}

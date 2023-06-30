package biz

import (
	"encoding/base64"
	"faceLiveDbus/pkg/goface"
	"io"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/patrickmn/go-cache"
)

const (
    CacheKeyMouse = "mouse"
    CacheKeyEye = "eye"
    CacheKeyImg = "img"
)

const FaceModels = "/home/omega/goprogram/faceLiveDbus/pkg/goface/"

const (
    CacheKeyDelay = 10 * time.Minute
    CacheDefaultExpiration = 1 * time.Hour
    CacheCleanupInterval = 2 * time.Hour
)

type File struct {
    io.ReadSeeker
    Size int64
}


type LiveDetected struct {
    Goface *goface.Goface
    Cache *cache.Cache
}

func (l LiveDetected) Init() (string, *dbus.Error) {
    l.Cache.Set(CacheKeyMouse, "", CacheKeyDelay)
    l.Cache.Set(CacheKeyEye, "", CacheKeyDelay)
    l.Cache.Set(CacheKeyImg, "", CacheKeyDelay) 
    return "done", nil
}


func (l LiveDetected) LiveDetection(detectedType string, b64Str string) (string, *dbus.Error) {
    base64Byte := make([]byte, base64.StdEncoding.DecodedLen(len(b64Str)))
    n, err := base64.StdEncoding.Decode(base64Byte, []byte(b64Str))
    if err != nil {
        return "", &dbus.Error{
            Name: "org.freedesktop.DBus.Error.InvalidArgs",
            Body: []interface{}{"image error: "+err.Error()},
        }
    }
    buf := base64Byte[:n]

    switch detectedType {
    case goface.MouseDetected:
        mousev, ok := l.IsMousePass()
        if ok {
            return "1", nil 
        }
        if !ok && mousev == "error" {
            return "0", &dbus.Error{
                Name: "org.freedesktop.DBus.Error.InvalidArgs",
                Body: []interface{}{"mouse cache not init"},
            }
        }
        mouseCode, err := l.Goface.Rec.MouseDetectedFromFile(buf)
        if err != nil {
            return "0", &dbus.Error{
                Name: "org.freedesktop.DBus.Error.InvalidArgs",
                Body: []interface{}{"mouse detected error: "+err.Error()},
            }
        }
        log.Printf("got mouse code: %v", mouseCode)
        if mouseCode != goface.MouseOpen && mouseCode != goface.MouseClose {
            return "0", &dbus.Error{
                Name: "org.freedesktop.DBus.Error.InvalidArgs",
                Body: []interface{}{"mouse detected fail"},
            }
        }
        l.Cache.Set(CacheKeyMouse, mousev+strconv.Itoa(mouseCode), CacheKeyDelay)
        if mouseCode == goface.MouseClose {
            err = l.saveFaceImg(b64Str)
            if err != nil {
                return "0", &dbus.Error{
                    Name: "org.freedesktop.DBus.Error.InvalidArgs",
                    Body: []interface{}{"detected face save error: "+err.Error()},
                }
            }
        }
    case goface.EyeDetected:
        eyev, ok := l.IsEyePass()
        if ok {
            return "1", nil 
        }
        if !ok && eyev == "error" {
            return "0", &dbus.Error{
                Name: "org.freedesktop.DBus.Error.InvalidArgs",
                Body: []interface{}{"eye cache not init"},
            }
        }
        eyeCode, err := l.Goface.Rec.EyeDetectedFromFile(buf)
        if err != nil {
            return "0", &dbus.Error{
                Name: "org.freedesktop.DBus.Error.InvalidArgs",
                Body: []interface{}{"eye detected error: "+err.Error()},
            }
        }
        if eyeCode != goface.EyeOpen && eyeCode != goface.EyeClose {
            return "0", &dbus.Error{
                Name: "org.freedesktop.DBus.Error.InvalidArgs",
                Body: []interface{}{"eye detected fail"},
            }
        }
        l.Cache.Set(CacheKeyEye, eyev+strconv.Itoa(eyeCode), CacheKeyDelay)
    default:
        return "0", &dbus.Error{
                Name: "org.freedesktop.DBus.Error.InvalidArgs",
                Body: []interface{}{"live detected error, type on match!"},
        }
    }
    return "0", &dbus.Error{
                Name: "org.freedesktop.DBus.Error.InvalidArgs",
                Body: []interface{}{"live detected fail"},
    } 
}

func (l *LiveDetected) IsMousePass() (string, bool) {
    mouse, found := l.Cache.Get(CacheKeyMouse)
    if !found {
        return "error", false 
    }
    mousev, ok := mouse.(string)
    if !ok {
        return "", false 
    }
    if len(mousev) > 3 {
        re := regexp.MustCompile("1(0)+1")

        found := re.FindAllString(mousev, -1)
        if found != nil && len(found) >= 2 {
            return mousev, true 
        }
    }

    return mousev, false 
}


func (l LiveDetected) IsEyePass() (string, bool) {
    eye, found := l.Cache.Get(CacheKeyEye)
    if !found {
        return "error", false
    } 
    eyev, ok := eye.(string)
    if !ok {
        return "", false 
    }
    if len(eyev) > 3 {
        re := regexp.MustCompile("1(0)+1")

        found := re.FindAllString(eyev, -1)
        if found != nil && len(found) >= 2 {
            return eyev, true 
        }
    }
    return eyev, false
}

func (l LiveDetected) GetComparedImage() (string, *dbus.Error) {
    img, found := l.Cache.Get(CacheKeyImg)
    if !found {
        return "", &dbus.Error{
            Name: "org.freedesktop.DBus.Error.InvalidArgs",     
            Body: []interface{}{"image data not init"},
        }
    }
    imgStr, ok := img.(string)
    if !ok {
        return "", &dbus.Error{
            Name: "org.freedesktop.DBus.Error.InvalidArgs",
            Body: []interface{}{"image data type error"},
        }
    }

    return imgStr, nil
}

func (l LiveDetected) saveFaceImg(b64Img string) error {
    l.Cache.Set(CacheKeyImg, b64Img, CacheKeyDelay) 
    return nil
}

package main

import (
	"faceLiveDbus/biz"
	"faceLiveDbus/pkg/goface"
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/patrickmn/go-cache"
)

/*
                <method name="Foo">
			<arg direction="out" type="s"/>
		</method>
		<method name="Sleep">
			<arg direction="in" type="u"/>
		</method>
*/

const intro = `
<node>
	<interface name="face.Detected">
                <method name="LiveDetection">
                    <arg name="detectedType" type="s" direction="in"/>
                    <arg name="buf" type="s" direction="in"/>
                    <arg type="s" direction="out"/>
                </method>
                <method name="Init">
                    <arg type="s" direction="out"/>
                </method>
                <method name="GetComparedImage">
                    <arg type="s" direction="out"/>
                </method>
	</interface>` + introspect.IntrospectDataString + `</node> `

/*type foo string

func (f foo) Foo() (string, *dbus.Error) {
	fmt.Println(f)
	return string(f), nil
}

func (f foo) Sleep(seconds uint) *dbus.Error {
	fmt.Println("Sleeping", seconds, "second(s)")
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}*/

func main() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	f := biz.LiveDetected{
            Goface: goface.NewGoface(biz.FaceModels),
            Cache: cache.New(biz.CacheDefaultExpiration, biz.CacheCleanupInterval),
        }
	conn.Export(f, "/face/Detected", "face.Detected")
	conn.Export(introspect.Introspectable(intro), "/face/Detected",
		"org.freedesktop.DBus.Introspectable")

	reply, err := conn.RequestName("face.Detected",
		dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "name already taken")
		os.Exit(1)
	}
	fmt.Println("Listening on face.Detected / /face/Detected ...")
	select {}
}

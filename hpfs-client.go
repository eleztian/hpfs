package main

import (
	"flag"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"

	"bs-2018/hpfs-client/p2p"
	"encoding/json"
	"bs-2018/hpfs-client/config"
	"context"
)

// Constants
const (
	APP_VERSION = "1.0.0.0"
	htmlAbout   = `Welcome on <b>Astilectron</b> demo!<br>
This is using the bootstrap and the bundler.`
)

// Vars
var (
	AppName string
	BuiltAt string
	debug   = flag.Bool("d", false, "enables the debug mode")
	w       *astilectron.Window
)

func init() {
	p2p.SetHost(config.AppConfig.HostUrl)
	p2p.SetServerUP(config.AppConfig.User.Username, config.AppConfig.User.Password)
	err := p2p.StartService(context.Background(), config.AppConfig.RPCPort)
	if err != nil {
		panic(err)
	}
}

func main() {
	// Init
	flag.Parse()
	astilog.FlagInit()
	// Run bootstrap
	astilog.Debugf("Running app built at %s", BuiltAt)

	if err := bootstrap.Run(bootstrap.Options{
		Asset:    Asset,
		AssetDir: AssetDir,
		AstilectronOptions: astilectron.Options{
			AppName:            AppName,
			AppIconDarwinPath:  "resources/icon.icns",
			AppIconDefaultPath: "resources/icon.png",
		},
		Debug:    *debug,
		Homepage: "login.html",
		MenuOptions: []*astilectron.MenuItemOptions{
			{
				Label: astilectron.PtrStr("Config"),
				OnClick: func(e astilectron.Event) (deleteListener bool) {
					sendMessage("config", nil)
					return
				},
			},
			{
				Label: astilectron.PtrStr("About"),
				OnClick: func(e astilectron.Event) (deleteListener bool) {
					if err := bootstrap.SendMessage(w, "about", htmlAbout, func(m *bootstrap.MessageIn) {
						// Unmarshal payload
						var s string
						if err := json.Unmarshal(m.Payload, &s); err != nil {
							astilog.Error(errors.Wrap(err, "unmarshaling payload failed"))
							return
						}
						astilog.Infof("About modal has been displayed and payload is %s!", s)
					}); err != nil {
						astilog.Error(errors.Wrap(err, "sending about event failed"))
					}
					return
				},
			},
			{
				Role: astilectron.MenuItemRoleToggleDevTools,
			},
		},
		OnWait: func(_ *astilectron.Astilectron, iw *astilectron.Window, _ *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
			w = iw
			go func() {
				updateShareDirs()
			}()
			return nil
		},
		MessageHandler: handleMessages,
		RestoreAssets:  RestoreAssets,
		WindowOptions: &astilectron.WindowOptions{
			BackgroundColor: astilectron.PtrStr("#333"),
			Center:          astilectron.PtrBool(true),
			Height:          astilectron.PtrInt(900),
			Width:           astilectron.PtrInt(1300),
		},
	}); err != nil {
		astilog.Fatal(errors.Wrap(err, "running bootstrap failed"))
	}
}

func sendMessage(name string, payload interface{}) {
	if err := bootstrap.SendMessage(w, name, payload); err != nil {
		astilog.Error(errors.Wrap(err, "sending "+name+" event failed"))
	}
}

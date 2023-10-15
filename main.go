package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bloat/config"
	"bloat/renderer"
	"bloat/service"
)

var (
	configFiles = []string{"bloat.conf", "/etc/bloat.conf"}
)

func errExit(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func main() {
	configFile := flag.String("f", "", "config file")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()

	if len(*configFile) > 0 {
		configFiles = []string{*configFile}
	}
	config, err := config.ParseFiles(configFiles)
	if err != nil {
		errExit(err)
	}

	if !config.IsValid() {
		errExit(errors.New("invalid config"))
	}

	templatesGlobPattern := filepath.Join(config.TemplatesPath, "*")
	renderer, err := renderer.NewRenderer(templatesGlobPattern)
	if err != nil {
		errExit(err)
	}

	s := service.NewService(config.ClientName, config.ClientScope,
		config.ClientWebsite, config.CustomCSS, config.SingleInstance,
		config.PostFormats, renderer)
	handler := service.NewHandler(s, *verbose, config.StaticDirectory)

	log.Println("listening on", config.ListenAddress)
	err = http.ListenAndServe(config.ListenAddress, handler)
	if err != nil {
		errExit(err)
	}
}

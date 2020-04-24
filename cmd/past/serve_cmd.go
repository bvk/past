// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bvk/past/msg"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var serveCmd = &cobra.Command{
	Use:   "serve [flags]",
	Short: "Runs the messaging backend as a HTTP server.",
	RunE:  cmdServe,
}

func init() {
	flags := serveCmd.Flags()
	flags.Int("port", 8080, "TCP Port to listen for the requests.")
	flags.String("www-dir", "", "Root directory to serve files over HTTP.")
}

func cmdServe(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	port, err := flags.GetInt("port")
	if err != nil {
		return xerrors.Errorf("could not get --port value: %w", err)
	}
	wwwDir, err := flags.GetString("www-dir")
	if err != nil {
		return xerrors.Errorf("could not get --www-dir value: %w", err)
	}
	if len(wwwDir) == 0 {
		return xerrors.Errorf("www directory cannot be empty: %w", os.ErrInvalid)
	}
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}

	handler, err := msg.New(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create message handler: %w", err)
	}

	h := http.NewServeMux()
	h.Handle("/", http.FileServer(http.Dir(wwwDir)))
	h.Handle("/api", handler)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: h,
	}
	return s.ListenAndServe()
}

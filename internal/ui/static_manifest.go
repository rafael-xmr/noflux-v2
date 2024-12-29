// SPDX-FileCopyrightText: Copyright The Noflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "github.com/fiatjaf/noflux/internal/ui"

import (
	"net/http"

	"github.com/fiatjaf/noflux/internal/http/request"
	"github.com/fiatjaf/noflux/internal/http/response/json"
	"github.com/fiatjaf/noflux/internal/http/route"
	"github.com/fiatjaf/noflux/internal/model"
)

func (h *handler) showWebManifest(w http.ResponseWriter, r *http.Request) {
	type webManifestShareTargetParams struct {
		URL  string `json:"url"`
		Text string `json:"text"`
	}

	type webManifestShareTarget struct {
		Action  string                       `json:"action"`
		Method  string                       `json:"method"`
		Enctype string                       `json:"enctype"`
		Params  webManifestShareTargetParams `json:"params"`
	}

	type webManifestIcon struct {
		Source  string `json:"src"`
		Sizes   string `json:"sizes"`
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
	}

	type webManifest struct {
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		ShortName       string                 `json:"short_name"`
		StartURL        string                 `json:"start_url"`
		Icons           []webManifestIcon      `json:"icons"`
		ShareTarget     webManifestShareTarget `json:"share_target"`
		Display         string                 `json:"display"`
		BackgroundColor string                 `json:"background_color"`
	}

	displayMode := "standalone"
	if request.IsAuthenticated(r) {
		user, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
		displayMode = user.DisplayMode
	}
	themeColor := model.ThemeColor(request.UserTheme(r), "light")
	manifest := &webManifest{
		Name:            "Noflux",
		ShortName:       "Noflux",
		Description:     "Minimalist Feed Reader",
		Display:         displayMode,
		StartURL:        route.Path(h.router, "login"),
		BackgroundColor: themeColor,
		Icons: []webManifestIcon{
			{Source: route.Path(h.router, "appIcon", "filename", "icon-120.png"), Sizes: "120x120", Type: "image/png", Purpose: "any"},
			{Source: route.Path(h.router, "appIcon", "filename", "maskable-icon-120.png"), Sizes: "120x120", Type: "image/png", Purpose: "maskable"},
		},
		ShareTarget: webManifestShareTarget{
			Action:  route.Path(h.router, "bookmarklet"),
			Method:  http.MethodGet,
			Enctype: "application/x-www-form-urlencoded",
			Params:  webManifestShareTargetParams{URL: "uri", Text: "text"},
		},
	}

	json.OK(w, r, manifest)
}

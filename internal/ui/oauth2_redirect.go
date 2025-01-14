// SPDX-FileCopyrightText: Copyright The Noflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "github.com/fiatjaf/noflux/internal/ui"

import (
	"log/slog"
	"net/http"

	"github.com/fiatjaf/noflux/internal/http/request"
	"github.com/fiatjaf/noflux/internal/http/response/html"
	"github.com/fiatjaf/noflux/internal/http/route"
	"github.com/fiatjaf/noflux/internal/oauth2"
	"github.com/fiatjaf/noflux/internal/ui/session"
)

func (h *handler) oauth2Redirect(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))

	provider := request.RouteStringParam(r, "provider")
	if provider == "" {
		slog.Warn("Invalid or missing OAuth2 provider")
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	authProvider, err := getOAuth2Manager(r.Context()).FindProvider(provider)
	if err != nil {
		slog.Error("Unable to initialize OAuth2 provider",
			slog.String("provider", provider),
			slog.Any("error", err),
		)
		html.Redirect(w, r, route.Path(h.router, "login"))
		return
	}

	auth := oauth2.GenerateAuthorization(authProvider.GetConfig())

	sess.SetOAuth2State(auth.State())
	sess.SetOAuth2CodeVerifier(auth.CodeVerifier())

	html.Redirect(w, r, auth.RedirectURL())
}

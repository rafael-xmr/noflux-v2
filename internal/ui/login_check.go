// SPDX-FileCopyrightText: Copyright The Noflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "github.com/fiatjaf/noflux/internal/ui"

import (
	"log/slog"
	"net/http"

	"github.com/fiatjaf/noflux/internal/config"
	"github.com/fiatjaf/noflux/internal/http/cookie"
	"github.com/fiatjaf/noflux/internal/http/request"
	"github.com/fiatjaf/noflux/internal/http/response/html"
	"github.com/fiatjaf/noflux/internal/http/route"
	"github.com/fiatjaf/noflux/internal/locale"
	"github.com/fiatjaf/noflux/internal/ui/form"
	"github.com/fiatjaf/noflux/internal/ui/session"
	"github.com/fiatjaf/noflux/internal/ui/view"
)

func (h *handler) checkLogin(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	if config.Opts.DisableLocalAuth() {
		slog.Warn("blocking local auth login attempt, local auth is disabled",
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
		)
		html.OK(w, r, view.Render("login"))
		return
	}

	authForm := form.NewAuthForm(r)
	view.Set("errorMessage", locale.NewLocalizedError("error.bad_credentials").Translate(request.UserLanguage(r)))
	view.Set("form", authForm)

	if validationErr := authForm.Validate(); validationErr != nil {
		translatedErrorMessage := validationErr.Translate(request.UserLanguage(r))
		slog.Warn("Validation error during login check",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", authForm.Username),
			slog.Any("error", translatedErrorMessage),
		)
		html.OK(w, r, view.Render("login"))
		return
	}

	if err := h.store.CheckPassword(authForm.Username, authForm.Password); err != nil {
		slog.Warn("Incorrect username or password",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", authForm.Username),
			slog.Any("error", err),
		)
		html.OK(w, r, view.Render("login"))
		return
	}

	sessionToken, userID, err := h.store.CreateUserSessionFromUsername(authForm.Username, r.UserAgent(), clientIP)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	slog.Info("User authenticated successfully with username/password",
		slog.Bool("authentication_successful", true),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
		slog.String("username", authForm.Username),
	)

	h.store.SetLastLogin(userID)

	user, err := h.store.UserByID(userID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)

	http.SetCookie(w, cookie.New(
		cookie.CookieUserSessionID,
		sessionToken,
		config.Opts.HTTPS,
		config.Opts.BasePath(),
	))

	html.Redirect(w, r, route.Path(h.router, user.DefaultHomePage))
}

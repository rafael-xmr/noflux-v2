// SPDX-FileCopyrightText: Copyright The Noflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "github.com/fiatjaf/noflux/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/fiatjaf/noflux/internal/http/request"
	"github.com/fiatjaf/noflux/internal/http/response/json"
	"github.com/fiatjaf/noflux/internal/model"
	"github.com/fiatjaf/noflux/internal/validator"
)

func (h *handler) currentUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, user)
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	var userCreationRequest model.UserCreationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&userCreationRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateUserCreationWithPassword(h.store, &userCreationRequest); validationErr != nil {
		json.BadRequest(w, r, validationErr.Error())
		return
	}

	user, err := h.store.CreateUser(&userCreationRequest)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, user)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	userID := request.RouteInt64Param(r, "userID")

	var userModificationRequest model.UserModificationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&userModificationRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	originalUser, err := h.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if originalUser == nil {
		json.NotFound(w, r)
		return
	}

	if !request.IsAdminUser(r) {
		if originalUser.ID != request.UserID(r) {
			json.Forbidden(w, r)
			return
		}

		if userModificationRequest.IsAdmin != nil && *userModificationRequest.IsAdmin {
			json.BadRequest(w, r, errors.New("only administrators can change permissions of standard users"))
			return
		}
	}

	cleanEnd := regexp.MustCompile(`(?m)\r\n\s*$`)
	if userModificationRequest.BlockFilterEntryRules != nil {
		*userModificationRequest.BlockFilterEntryRules = cleanEnd.ReplaceAllLiteralString(*userModificationRequest.BlockFilterEntryRules, "")
		// Clean carriage returns for Windows environments
		*userModificationRequest.BlockFilterEntryRules = strings.ReplaceAll(*userModificationRequest.BlockFilterEntryRules, "\r\n", "\n")
	}
	if userModificationRequest.KeepFilterEntryRules != nil {
		*userModificationRequest.KeepFilterEntryRules = cleanEnd.ReplaceAllLiteralString(*userModificationRequest.KeepFilterEntryRules, "")
		// Clean carriage returns for Windows environments
		*userModificationRequest.KeepFilterEntryRules = strings.ReplaceAll(*userModificationRequest.KeepFilterEntryRules, "\r\n", "\n")
	}

	if validationErr := validator.ValidateUserModification(h.store, originalUser.ID, &userModificationRequest); validationErr != nil {
		json.BadRequest(w, r, validationErr.Error())
		return
	}

	userModificationRequest.Patch(originalUser)
	if err = h.store.UpdateUser(originalUser); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, originalUser)
}

func (h *handler) markUserAsRead(w http.ResponseWriter, r *http.Request) {
	userID := request.RouteInt64Param(r, "userID")
	if userID != request.UserID(r) {
		json.Forbidden(w, r)
		return
	}

	if _, err := h.store.UserByID(userID); err != nil {
		json.NotFound(w, r)
		return
	}

	if err := h.store.MarkAllAsRead(userID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) getIntegrationsStatus(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	if _, err := h.store.UserByID(userID); err != nil {
		json.NotFound(w, r)
		return
	}

	hasIntegrations := h.store.HasSaveEntry(userID)

	response := struct {
		HasIntegrations bool `json:"has_integrations"`
	}{
		HasIntegrations: hasIntegrations,
	}

	json.OK(w, r, response)
}

func (h *handler) users(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	users, err := h.store.Users()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	users.UseTimezone(request.UserTimezone(r))
	json.OK(w, r, users)
}

func (h *handler) userByID(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	user, err := h.store.UserByID(userID)
	if err != nil {
		json.BadRequest(w, r, errors.New("unable to fetch this user from the database"))
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	user.UseTimezone(request.UserTimezone(r))
	json.OK(w, r, user)
}

func (h *handler) userByUsername(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	username := request.RouteStringParam(r, "username")
	user, err := h.store.UserByUsername(username)
	if err != nil {
		json.BadRequest(w, r, errors.New("unable to fetch this user from the database"))
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	json.OK(w, r, user)
}

func (h *handler) removeUser(w http.ResponseWriter, r *http.Request) {
	if !request.IsAdminUser(r) {
		json.Forbidden(w, r)
		return
	}

	userID := request.RouteInt64Param(r, "userID")
	user, err := h.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if user == nil {
		json.NotFound(w, r)
		return
	}

	if user.ID == request.UserID(r) {
		json.BadRequest(w, r, errors.New("you cannot remove yourself"))
		return
	}

	h.store.RemoveUserAsync(user.ID)
	json.NoContent(w, r)
}

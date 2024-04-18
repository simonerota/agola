// Copyright 2019 Sorint.lab
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/sorintlab/errors"

	"agola.io/agola/internal/services/configstore/action"
	"agola.io/agola/internal/util"
	csapitypes "agola.io/agola/services/configstore/api/types"
)

type MaintenanceStatusHandler struct {
	log                    zerolog.Logger
	ah                     *action.ActionHandler
	currentMaintenanceMode bool
}

func NewMaintenanceStatusHandler(log zerolog.Logger, ah *action.ActionHandler, currentMaintenanceMode bool) *MaintenanceStatusHandler {
	return &MaintenanceStatusHandler{log: log, ah: ah, currentMaintenanceMode: currentMaintenanceMode}
}

func (h *MaintenanceStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.do(r)
	if util.HTTPError(w, err) {
		h.log.Err(err).Send()
		return
	}

	if err := util.HTTPResponse(w, http.StatusOK, res); err != nil {
		h.log.Err(err).Send()
	}
}

func (h *MaintenanceStatusHandler) do(r *http.Request) (*csapitypes.MaintenanceStatusResponse, error) {
	ctx := r.Context()

	requestedStatus, err := h.ah.IsMaintenanceEnabled(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := &csapitypes.MaintenanceStatusResponse{RequestedStatus: requestedStatus, CurrentStatus: h.currentMaintenanceMode}

	return res, nil
}

type MaintenanceModeHandler struct {
	log zerolog.Logger
	ah  *action.ActionHandler
}

func NewMaintenanceModeHandler(log zerolog.Logger, ah *action.ActionHandler) *MaintenanceModeHandler {
	return &MaintenanceModeHandler{log: log, ah: ah}
}

func (h *MaintenanceModeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.do(r)
	if util.HTTPError(w, err) {
		h.log.Err(err).Send()
		return
	}
}

func (h *MaintenanceModeHandler) do(r *http.Request) error {
	ctx := r.Context()

	enable := false
	switch r.Method {
	case "PUT":
		enable = true
	case "DELETE":
		enable = false
	}

	err := h.ah.SetMaintenanceEnabled(ctx, enable)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type ExportHandler struct {
	log zerolog.Logger
	ah  *action.ActionHandler
}

func NewExportHandler(log zerolog.Logger, ah *action.ActionHandler) *ExportHandler {
	return &ExportHandler{log: log, ah: ah}
}

func (h *ExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.do(w, r)
	if util.HTTPError(w, err) {
		h.log.Err(err).Send()
		return
	}

}

func (h *ExportHandler) do(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if err := h.ah.Export(ctx, w); err != nil {
		h.log.Err(err).Send()
		// since we already answered with a 200 we cannot return another error code
		// So abort the connection and the client will detect the missing ending chunk
		// and consider this an error
		//
		// this is the way to force close a request without logging the panic
		panic(http.ErrAbortHandler)
	}

	return nil
}

type ImportHandler struct {
	log zerolog.Logger
	ah  *action.ActionHandler
}

func NewImportHandler(log zerolog.Logger, ah *action.ActionHandler) *ImportHandler {
	return &ImportHandler{log: log, ah: ah}
}

func (h *ImportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.do(r)
	if util.HTTPError(w, err) {
		h.log.Err(err).Send()
		return
	}
}

func (h *ImportHandler) do(r *http.Request) error {
	ctx := r.Context()

	if err := h.ah.Import(ctx, r.Body); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Package ui provides the main entry point for the UI.
package ui

import (
	"github.com/new985211/fight-the-landlord/internal/ui/handler"
	"github.com/new985211/fight-the-landlord/internal/ui/model"
	"github.com/new985211/fight-the-landlord/internal/ui/view"
)

// NewOnlineModel creates a new OnlineModel for online game mode.
func NewOnlineModel(serverURL string) *model.OnlineModel {
	m := model.NewOnlineModel(serverURL)
	m.SetViewRenderer(view.CreateViewRenderer())
	m.SetKeyHandler(HandleKeyPress)
	m.SetServerMessageHandler(handler.HandleServerMessage)
	return m
}

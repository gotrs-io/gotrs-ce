package api

import "github.com/gotrs-io/gotrs-ce/internal/routing"

func init() {
	// Agent template API endpoints
	routing.GlobalHandlerMap["handleGetAgentTemplates"] = handleGetAgentTemplates
	routing.GlobalHandlerMap["handleGetAgentTemplate"] = handleGetAgentTemplate
}

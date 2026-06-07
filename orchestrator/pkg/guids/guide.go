package guids

import "orchestrator/pkg/guids/core"

// CommandExample — см. core.CommandExample.
type CommandExample = core.CommandExample

// Guide — см. core.Guide.
type Guide = core.Guide

// Get возвращает Guide по имени инструмента. Делегирует в core.Get.
func Get(name string) *Guide {
	return core.Get(name)
}

// All возвращает все зарегистрированные Guide'ы.
func All() []*Guide {
	return core.All()
}

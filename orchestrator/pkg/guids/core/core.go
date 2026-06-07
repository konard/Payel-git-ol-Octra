package core

import (
	"fmt"
	"strings"
)

// CommandExample документирует одну типовую команду для AI-агента.
type CommandExample struct {
	Purpose string
	Command string
}

// Guide — структурированная шпаргалка по одному инструменту/фреймворку.
type Guide struct {
	Name      string
	Tool      string
	Tools     []string
	Desc      string
	Commands  []CommandExample
	Structure string
}

var registry []*Guide

// Register добавляет Guide в глобальный реестр.
func Register(g Guide) {
	registry = append(registry, &g)
}

// Get возвращает Guide по имени инструмента (регистронезависимо).
func Get(name string) *Guide {
	for _, g := range registry {
		if strings.EqualFold(g.Name, name) || strings.EqualFold(g.Tool, name) {
			return g
		}
	}
	return nil
}

// All возвращает все зарегистрированные Guide'ы.
func All() []*Guide {
	return registry
}

// Format возвращает однострочник с кратким описанием (используется в ManagerThink).
func (g *Guide) FormatBrief() string {
	return g.Desc
}

// FormatCheat возвращает полную шпаргалку для вставки в Worker-промпты:
//
//	TOOL REFERENCE — <Name>:
//	  <Purpose>: <Command>
//	  ...
//	  Typical structure:
//	  <Structure>
func (g *Guide) FormatCheat() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("TOOL REFERENCE — %s:\n", g.Name))
	for _, cmd := range g.Commands {
		b.WriteString(fmt.Sprintf("  %s: %s\n", cmd.Purpose, cmd.Command))
	}
	if g.Structure != "" {
		b.WriteString("\nTypical project structure:\n")
		b.WriteString(g.Structure)
	}
	return b.String()
}

// ShortDesc возвращает краткое описание инструмента (одна строка).
func (g *Guide) ShortDesc() string {
	names := make([]string, 0, len(g.Commands))
	for _, cmd := range g.Commands {
		names = append(names, cmd.Command)
	}
	return strings.Join(names, "  /  ")
}

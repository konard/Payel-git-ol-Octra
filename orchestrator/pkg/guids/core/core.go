package core

import "strings"

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

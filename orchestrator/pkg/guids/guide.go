package guids

import "strings"

// CommandExample документирует одну типовую команду для AI-агента.
type CommandExample struct {
	Purpose string // что делает команда (на английском, для промпта)
	Command string // сама команда (cargo init, dotnet new, ...)
}

// Guide — структурированная шпаргалка по одному инструменту/фреймворку.
// Используется AI-агентами (Boss → Manager → Worker) чтобы не тратить
// токены на угадывание базовых команд.
type Guide struct {
	Name      string            // отображаемое имя: "Cargo", "Composer"
	Tool      string            // имя бинарника: "cargo", "composer"
	Tools     []string          // Nix-пакеты, которые нужны: ["rustc", "cargo"]
	Desc      string            // краткое описание
	Commands  []CommandExample  // упорядоченный список типовых команд
	Structure string            // типичная структура проекта
}

var registry []*Guide

func register(g Guide) {
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

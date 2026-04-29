# CrewAI Chat Context - 2026-04-27

## Current State
- CrewAI система многоагентной генерации кода работает
- Git интеграция полностью реализована
- ZIP удалён, используется только GitHub

## Recent Changes
1. **Добавлено поле GitData в модель Task** (boss/pkg/models/task.go)
2. **Добавлены функции SaveGitData/RestoreGitData** в boss/manager/worker
3. **Сохранение .git при завершении задачи** (boss/task.go)
4. **Восстановление .git + commit при restore** (boss/boss.go)

## Technical Notes
- GitHub токен: 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
- Git workflow: boss → manager branches → worker commits → manager merges → boss pushes
- При сохранении: .git папка сериализуется в JSON и сохраняется в БД
- При восстановлении: .git восстанавливается, файлы создаются, делается git add + commit

## Build Status
- ✅ boss: OK
- ✅ manager: OK  
- ✅ worker: OK</content>
<parameter name="filePath">.claude/context/crewai-git-integration-context.md
package rules

// Этот пакет описывает внутренние контракты между узлами boss → manager → worker.
// Ранее эти контракты были описаны в .proto и шли по gRPC; теперь это обычные
// Go-структуры и прямые вызовы внутри одного процесса.

// WorkerRole — описание роли воркера, который менеджер хочет нанять
type WorkerRole struct {
	Role         string
	Description  string
	CustomPrompt string
}

// ManagerRole — описание роли менеджера
type ManagerRole struct {
	Role         string
	Description  string
	Priority     int32
	CustomPrompt string
}

// AssignManagerRequest — задание для одного менеджера от босса
type AssignManagerRequest struct {
	TaskId               string
	ManagerId            string
	Role                 string
	Description          string
	CustomPrompt         string
	TechnicalDescription string
	ProjectPath          string
	Metadata             map[string]string
	OtherWorkersResults  []*WorkerResult
}

// AssignWorkersRequest — задание для группы воркеров от менеджера
type AssignWorkersRequest struct {
	TaskId              string
	ManagerId           string
	ManagerRole         string
	WorkerRoles         []*WorkerRole
	TaskMd              string
	ProjectPath         string
	Metadata            map[string]string
	OtherWorkersResults []*WorkerResult
}

// WorkerResult — результат работы одного воркера
type WorkerResult struct {
	WorkerId   string
	Role       string
	TaskMd     string
	SolutionMd string
	Files      map[string]string
	Success    bool
	Approved   bool
	Feedback   string
	Commands   []string
}

// AssignWorkersResponse — итог работы группы воркеров
type AssignWorkersResponse struct {
	TaskId        string
	Status        string
	WorkerResults []*WorkerResult
}

// ManagerResult — итог работы одного менеджера
type ManagerResult struct {
	TaskId        string
	ManagerId     string
	Role          string
	Status        string
	WorkerResults []*WorkerResult
	ReviewSummary string
}

// ReviewRequest — менеджер просит воркера исправить решение
type ReviewRequest struct {
	TaskId             string
	ManagerId          string
	WorkerId           string
	WorkerRole         string
	Feedback           string
	OriginalSolutionMd string
	OriginalFiles      map[string]string
	Metadata           map[string]string
}

// ReviewResponse — результат исправления от воркера
type ReviewResponse struct {
	TaskId     string
	WorkerId   string
	Status     string // "fixed" | "failed"
	Feedback   string
	SolutionMd string
	Files      map[string]string
}

// ProgressFunc — callback для пробрасывания прогресса вверх по цепочке
type ProgressFunc func(progress int32, message string, data map[string]string)

package cache

const (
	taskKeyPrefix = "tasks:task:"
	listKey       = "tasks:list"
)

// TaskKey генерирует ключ для кэширования задачи по ID
func TaskKey(id string) string {
	return taskKeyPrefix + id
}

// ListKey генерирует ключ для кэширования списка задач
func ListKey() string {
	return listKey
}

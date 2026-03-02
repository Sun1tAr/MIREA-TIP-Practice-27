package cache

import (
	"math/rand"
	"time"
)

// GetTTL возвращает TTL с джиттером для предотвращения cache avalanche
func GetTTL(baseTTL int, jitter int) time.Duration {
	if jitter <= 0 {
		return time.Duration(baseTTL) * time.Second
	}

	// Генерируем случайное значение от -jitter до +jitter
	// Чтобы не уходить в отрицательные значения, делаем от 0 до jitter*2
	// и вычитаем jitter
	jitterValue := rand.Intn(jitter*2) - jitter

	ttl := baseTTL + jitterValue
	if ttl <= 0 {
		ttl = 1 // минимальный TTL 1 секунда
	}

	return time.Duration(ttl) * time.Second
}

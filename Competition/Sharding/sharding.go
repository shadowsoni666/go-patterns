package Sharding

import (
	"crypto/sha1"
	"fmt"
	"sync"
)

// Внутренне ShardedMap – это самый обычный срез с указателями на экземпляры Shard, но мы определяем его как отдельный тип, чтобы получить возможность присоединять к нему методы. Каждый экземпляр Shard включает
// поле map[string]interface{}, содержащее информацию о сегменте, и скомпонован с  sync.RWMutex, чтобы дать возможность заблокировать этот сегмент
// отдельно:
type Shard struct {
	sync.RWMutex                        // Встраивание sync.RWMutex
	m            map[string]interface{} // m содержит информацию о сегменте
}
type ShardedMap []*Shard // ShardedMap является срезом с экземплярами *Shard

// В языке Go нет понятия «конструктор», поэтому добавим функцию NewShardedMap для создания нового экземпляра ShardedMap
func NewShardedMap(nshards int) ShardedMap {
	shards := make([]*Shard, nshards) // Инициализировать срез с экземплярами *Shard
	for i := 0; i < nshards; i++ {
		shard := make(map[string]interface{})
		shards[i] = &Shard{m: shard}
	}
	return shards // ShardedMap ЯВЛЯЕТСЯ срезом с экземплярами *Shard!
}

//ShardedMap имеет два внутренних метода, getShardIndex и getShard, которые
//используются для вычисления индекса сегмента по ключу и получения соответствующего сегмента.
//Их можно объединить в один метод, но такое разделение, как здесь, упрощает их тестировани

func (m ShardedMap) getShardIndex(key string) int {
	checksum := sha1.Sum([]byte(key)) // Использовать Sum из "crypto/sha1"
	hash := int(checksum[17])         // Выбрать произвольный байт на роль хеша
	return hash % len(m)              // Взять остаток от деления на len(m),
	// чтобы получить индекс
}
func (m ShardedMap) getShard(key string) *Shard {
	index := m.getShardIndex(key)
	return m[index]
}

func (m ShardedMap) Get(key string) interface{} {
	shard := m.getShard(key)
	shard.RLock()
	defer shard.RUnlock()
	return shard.m[key]
}
func (m ShardedMap) Set(key string, value interface{}) {
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.m[key] = value
}

// Если понадобится установить блокировки для всех сегментов, то лучше
// делать это одновременно.
// Ниже показана реализация функции Keys с использованием сопрограмм и нашего старого знакомого sync.WaitGroup:
func (m ShardedMap) Keys() []string {
	keys := make([]string, 0) // Создать пустой срез ключей
	mutex := sync.Mutex{}     // Мьютекс для безопасной записи в keys
	wg := sync.WaitGroup{}    // Создать группу ожидания и установить
	wg.Add(len(m))            // счетчик равным количеству сегментов
	for _, shard := range m { // Запустить сопрограмму для каждого сегмента
		go func(s *Shard) {
			s.RLock()              // Установить блокировку для чтения в s
			for key := range s.m { // Получить ключи из сегмента
				mutex.Lock()
				keys = append(keys, key)
				mutex.Unlock()
			}
			s.RUnlock() // Снять блокировку для чтения
			wg.Done()   // Сообщить WaitGroup, что обработка завершена
		}(shard)
	}
	wg.Wait() // Приостановить выполнение до выполнения
	// всех операций чтения
	return keys // Вернуть срез с ключами
}

func (m ShardedMap) Delete(key string) {

}

func main() {
	shardedMap := NewShardedMap(5)
	shardedMap.Set("alpha", 1)
	shardedMap.Set("beta", 2)
	shardedMap.Set("gamma", 3)
	fmt.Println(shardedMap.Get("alpha"))
	fmt.Println(shardedMap.Get("beta"))
	fmt.Println(shardedMap.Get("gamma"))
	keys := shardedMap.Keys()
	for _, k := range keys {
		fmt.Println(k)
	}
}

package cache

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"

	"telematics-generator/pkg/models"
)

type DataCacher interface {
	GetLatest() (models.TelematicsData, bool)
	GetRange(time.Time, time.Time) ([]models.TelematicsData, error)
	Add(models.TelematicsData)
}

type TelematicsDataCache struct {
	capacity     int
	data         map[time.Time]*list.Element
	list         *list.List
	mx           sync.Mutex
	minTimestamp time.Time
	maxTimestamp time.Time
}

type Item struct {
	Timestamp time.Time
	Data      models.TelematicsData
}

func NewTelematicsDataCache(capacity int) *TelematicsDataCache {
	return &TelematicsDataCache{
		capacity:     capacity,
		data:         make(map[time.Time]*list.Element),
		list:         list.New(),
		minTimestamp: time.Now(),
		maxTimestamp: time.Now(),
	}
}

func (c *TelematicsDataCache) Add(telematicsData models.TelematicsData) {
	c.mx.Lock()
	defer c.mx.Unlock()

	item := Item{
		Timestamp: telematicsData.Timestamp,
		Data:      telematicsData,
	}
	element := c.list.PushFront(item)
	c.data[item.Timestamp] = element

	if telematicsData.Timestamp.Before(c.minTimestamp) {
		c.minTimestamp = telematicsData.Timestamp
	}
	if telematicsData.Timestamp.After(c.maxTimestamp) {
		c.maxTimestamp = telematicsData.Timestamp
	}
}

func (c *TelematicsDataCache) GetLatest() (models.TelematicsData, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.list.Len() == 0 {
		return models.TelematicsData{}, false
	}

	latest := c.list.Front()
	return latest.Value.(Item).Data, true
}

func (c *TelematicsDataCache) GetRange(from, to time.Time) ([]models.TelematicsData, error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if from.After(to) {
		return nil, errors.New("the 'from' timestamp must be earlier than the 'to' timestamp")
	}
	if from.After(c.maxTimestamp) || to.Before(c.minTimestamp) {
		return nil, fmt.Errorf("requested range is out of bounds. Available range is from %v (timestamp: %v) to %v (timestamp: %v)",
			c.minTimestamp, c.minTimestamp.UnixNano(), c.maxTimestamp, c.maxTimestamp.UnixNano())

	}

	var result []models.TelematicsData
	for e := c.list.Front(); e != nil; e = e.Next() {
		item := e.Value.(Item)
		if item.Timestamp.After(from) && item.Timestamp.Before(to) {
			result = append(result, item.Data)
		}
	}

	return result, nil
}

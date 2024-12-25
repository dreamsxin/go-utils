package stats

import (
	"fmt"
	"strings"
)

// Stats 结构体用于存储各种统计信息。
type Stats struct {
	Scope          Namespace              `json:"scope"`           // 统计命名空间
	Name           string                 `json:"name"`            // 统计名称
	ID             uint32                 `json:"id"`              // 统计ID
	ExecutionStats map[string]interface{} `json:"execution_stats"` // 总执行次数
	FailureStats   map[string]interface{} `json:"failure_stats"`   // 失败次数

	Insight          *Insight        `json:"-"` // 运行效能信息
	LatencyHistogram *HistogramStats `json:"-"` // 延迟直方图统计信息
}

// 创建新的统计对象
func NewStats(statsInit bool, scope Namespace, name string, id uint32) *Stats {
	newStats := &Stats{
		Scope:            scope,
		Name:             name,
		ID:               id,
		ExecutionStats:   make(map[string]interface{}),
		FailureStats:     make(map[string]interface{}),
		Insight:          NewInsight(),
		LatencyHistogram: NewHistogramStats(),
	}

	return newStats
}

func (s *Stats) String() string {
	var stringBuilder strings.Builder

	stringBuilder.Grow(2048)
	stringBuilder.WriteString("{ \"execution_stats\": {")
	first := true
	for eStatField, eStatValue := range s.ExecutionStats {
		if !first {
			stringBuilder.WriteRune(',')
		}
		stringBuilder.WriteString(eStatField)
		stringBuilder.WriteRune(':')
		stringBuilder.WriteString(fmt.Sprintf("%v", eStatValue))
		first = false
	}

	stringBuilder.WriteString("}, \"failure_stats\" : {")
	first = true
	for fStatField, fStatValue := range s.FailureStats {
		if !first {
			stringBuilder.WriteRune(',')
		}
		stringBuilder.WriteString(fStatField)
		stringBuilder.WriteRune(':')
		stringBuilder.WriteString(fmt.Sprintf("%v", fStatValue))
		first = false
	}
	stringBuilder.WriteRune('}')

	return stringBuilder.String()
}

package stats

// 运行效能信息
type InsightLine struct {
	CallCount      int64   `json:"call_count"`
	CallTime       float64 `json:"call_time"`
	ExceptionCount int64   `json:"error_count"`
	LastException  string  `json:"error_msg"`
	LastLog        string  `json:"last_log"`
}

// 每一行运行效能信息
type Insight struct {
	Script string              `json:"script"`
	Lines  map[int]InsightLine `json:"lines"`
}

type Insights map[string]*Insight

func NewInsight() *Insight {
	return &Insight{Lines: make(map[int]InsightLine)}
}

func NewInsights() *Insights {
	o := make(Insights)
	return &o
}

func (dst *Insights) Accumulate(src *Insights) {
	for app, insight := range *src {
		val := (*dst)[app]
		if val == nil {
			val = NewInsight()
		}
		val.Accumulate(insight)
		(*dst)[app] = val
	}
}

func (dst *Insight) Accumulate(src *Insight) {
	for line, right := range src.Lines {
		left := dst.Lines[line]
		left.CallCount += right.CallCount
		left.CallTime += right.CallTime
		left.ExceptionCount += right.ExceptionCount
		if len(right.LastException) > 0 {
			left.LastException = right.LastException
		}
		if len(right.LastLog) > 0 {
			left.LastLog = right.LastLog
		}
		dst.Lines[line] = left
	}
	if len(src.Script) > 0 {
		dst.Script = src.Script
	}
}

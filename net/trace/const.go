package trace

// Trace key
const (
	// 历史遗留的 key 不要轻易修改
	KeyTraceID       = "x1-et-id"
	KeyTraceSpanID   = "x1-et-spanid"
	KeyTraceParentID = "x1-et-parentid"
	KeyTraceSampled  = "x1-et-sampled"
	KeyTraceLevel    = "x1-et-lv"
	KeyTraceCaller   = "x1-et-user"
	// trace sdk should use jd100_trace_id to get trace info after this code be merged
	ETTraceID    = "et-trace-id"
	ETTraceDebug = "et-trace-debug"
)

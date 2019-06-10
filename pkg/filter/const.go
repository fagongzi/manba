package filter

const (
	// AttrClientRealIP real client ip
	AttrClientRealIP = "__internal_real_ip__"
	// AttrUsingCachingValue using cached value to response
	AttrUsingCachingValue = "__internal_using_cache_value__"
	// AttrUsingResponse using response to response
	AttrUsingResponse = "__internal_using_response__"

	// BreakFilterChainCode break filter chain code
	BreakFilterChainCode = -1
)

// StringValue returns the attr value
func StringValue(attr string, c Context) string {
	return c.GetAttr(attr).(string)
}

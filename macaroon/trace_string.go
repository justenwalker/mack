// Code generated by "stringer -type=TraceOpKind -linecomment -output trace_string.go"; DO NOT EDIT.

package macaroon

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TraceOpUnknown-0]
	_ = x[TraceOpHMAC-1]
	_ = x[TraceOpDecrypt-2]
	_ = x[TraceOpBind-3]
	_ = x[TraceOpFail-4]
}

const _TraceOpKind_name = "UnknownHMACDecryptBindForRequestFAILURE"

var _TraceOpKind_index = [...]uint8{0, 7, 11, 18, 32, 39}

func (i TraceOpKind) String() string {
	if i < 0 || i >= TraceOpKind(len(_TraceOpKind_index)-1) {
		return "TraceOpKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TraceOpKind_name[_TraceOpKind_index[i]:_TraceOpKind_index[i+1]]
}
// Code generated by "stringer -type=Key"; DO NOT EDIT.

package tomato

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Left-0]
	_ = x[Right-1]
	_ = x[Up-2]
	_ = x[Down-3]
	_ = x[Escape-4]
	_ = x[Space-5]
	_ = x[Backspace-6]
	_ = x[Delete-7]
	_ = x[Enter-8]
	_ = x[Tab-9]
	_ = x[Home-10]
	_ = x[End-11]
	_ = x[PageUp-12]
	_ = x[PageDown-13]
	_ = x[Shift-14]
	_ = x[Ctrl-15]
	_ = x[Alt-16]
}

const _Key_name = "LeftRightUpDownEscapeSpaceBackspaceDeleteEnterTabHomeEndPageUpPageDownShiftCtrlAlt"

var _Key_index = [...]uint8{0, 4, 9, 11, 15, 21, 26, 35, 41, 46, 49, 53, 56, 62, 70, 75, 79, 82}

func (i Key) String() string {
	if i >= Key(len(_Key_index)-1) {
		return "Key(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Key_name[_Key_index[i]:_Key_index[i+1]]
}

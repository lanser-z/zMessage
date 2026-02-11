package dal

import "errors"

var (
	// ErrNotFound 记录不存在
	ErrNotFound = errors.New("record not found")

	// ErrDuplicate 唯一约束冲突
	ErrDuplicate = errors.New("duplicate record")

	// ErrInvalidArgument 无效参数
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrInternal 内部错误
	ErrInternal = errors.New("internal error")
)

// IsDuplicate 判断是否为唯一约束冲突
func IsDuplicate(err error) bool {
	return err == ErrDuplicate
}

// IsNotFound 判断是否为记录不存在
func IsNotFound(err error) bool {
	return err == ErrNotFound
}

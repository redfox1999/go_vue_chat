package utils

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
)

// F 类似 Python 的 f-string 格式化函数
// 使用方式: F("Hello {name}, you are {age} years old", map[string]any{"name": "Alice", "age": 25})
// 或: F("User {ID} is {Name}", userStruct)
func F(format string, data any) string {
	re := regexp.MustCompile(`\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(format, func(match string) string {
		// 去除大括号
		key := match[1 : len(match)-1]

		var value any

		// 尝试从 map 获取
		if m, ok := data.(map[string]any); ok {
			value = m[key]
		} else if m, ok := data.(map[string]string); ok {
			value = m[key]
		} else if m, ok := data.(map[string]int); ok {
			value = m[key]
		} else {
			// 尝试从结构体字段获取
			v := reflect.ValueOf(data)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			if v.Kind() == reflect.Struct {
				field := v.FieldByName(key)
				if field.IsValid() && field.CanInterface() {
					value = field.Interface()
				}
			}
		}

		if value == nil {
			return match // 保持原样
		}

		return fmt.Sprintf("%v", value)
	})
}

// Fprintf 类似 fmt.Fprintf，但使用 f-string 语法
func Fprintf(w io.Writer, format string, data any) (n int, err error) {
	result := F(format, data)
	return w.Write([]byte(result))
}

// Fprintln 类似 fmt.Fprintln，但使用 f-string 语法
func Fprintln(w io.Writer, format string, data any) (n int, err error) {
	result := F(format, data) + "\n"
	return w.Write([]byte(result))
}

// SprintF 类似 fmt.Sprintf，但使用 f-string 语法 (F 的别名)
func SprintF(format string, data any) string {
	return F(format, data)
}

// 示例用法:
// func main() {
//     // 使用 map
//     msg1 := F("Hello {name}, you are {age} years old", map[string]any{
//         "name": "Alice",
//         "age": 25,
//     })
//     fmt.Println(msg1) // Hello Alice, you are 25 years old
//
//     // 使用结构体
//     type User struct {
//         ID   int
//         Name string
//     }
//     user := User{ID: 1, Name: "Bob"}
//     msg2 := F("User {ID} is {Name}", user)
//     fmt.Println(msg2) // User 1 is Bob
// }

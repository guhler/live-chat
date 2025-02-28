package templ

import (
	"errors"
	"html/template"
	"reflect"

	"github.com/labstack/echo/v4"
)

func InitTempl(e *echo.Echo) {

	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict key must be string")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"arr": func(values ...any) ([]any, error) {
			return values, nil
		},
		"hasfield": func(v any, name string) bool {
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() != reflect.Struct {
				return false
			}
			return rv.FieldByName(name).IsValid()
		},
	}).ParseGlob("../templates/*.html"))

	e.Renderer = &echo.TemplateRenderer{
		Template: tmpl,
	}
}

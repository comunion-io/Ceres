// @EgoctlOverwrite YES
// @EgoctlGenerateTime 20210223_202936
package core

import (
	"net/http"
	"github.com/gin-gonic/gin"
    "errors"
    "reflect"
    "strings"
    "sync"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/locales/zh"
    ut "github.com/go-playground/universal-translator"
    "github.com/go-playground/validator/v10"
    tzh "github.com/go-playground/validator/v10/translations/zh"
    "github.com/gotomicro/ego/core/elog"
)



func init() {
	binding.Validator = &defaultValidator{}
}

// HandlerFunc core封装后的handler
type HandlerFunc func(c *Context)

// Handle 将core.HandlerFunc转换为gin.HandlerFunc
func Handle(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context{
			c,
		}
		h(ctx)
	}
}

// Context core封装后的Context
type Context struct {
	*gin.Context
}

const (
	// CodeOK 表示响应成功状态码
	CodeOK = 0
	// CodeErr 表示默认响应失败状态码
	CodeErr = 1
)

// Res 标准JSON输出格式
type Res struct {
	// Code 响应的业务错误码。0表示业务执行成功，非0表示业务执行失败。
	Code int `json:"code"`
	// Msg 响应的参考消息。前端可使用msg来做提示
	Msg string `json:"msg"`
	// Data 响应的具体数据
	Data interface{} `json:"data"`
}

// ResPage 带分页的标准JSON输出格式
type ResPage struct {
	Res
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	// Current 总记录数
	Current int `json:"current"`
	// PageSize 每页记录数
	PageSize int `json:"pageSize"`
	// Total 总页数
	Total int `json:"total"`
}

// JSON 输出响应JSON
// 形如 {"code":<code>, "msg":<msg>, "data":<data>}
func (c *Context) JSON(httpStatus int, res Res) {
	c.Context.JSON(httpStatus, res)
}

// JSONOK 输出响应成功JSON，如果data不为零值，则输出data
// 形如 {"code":0, "msg":"成功", "data":<data>}
func (c *Context) JSONOK(data ...interface{}) {
	j := new(Res)
	j.Code = CodeOK
	j.Msg = "成功"
	if len(data) > 0 {
		j.Data = data[0]
	} else {
		j.Data = ""
	}
	c.Context.JSON(http.StatusOK, j)
	return
}

// JSONE 输出失败响应
// 形如 {"code":<code>, "msg":<msg>, "data":<data>}
func (c *Context) JSONE(code int, msg string, data error) {
	j := new(Res)
	j.Code = code
	j.Msg = msg
	if data != nil {
		j.Data = data.Error()
	}
	c.Context.JSON(http.StatusOK, j)
	return
}

// JSONPage 输出带分页信息的JSON
// 形如 {"code":<code>, "msg":<msg>, "data":<data>, "pagination":<pagination>}
// <pagination> { "current":1, "pageSize":20, "total": 9 }
func (c *Context) JSONPage(data interface{}, pagination Pagination) {
	j := new(ResPage)
	j.Code = CodeOK
	j.Data = data
	j.Pagination = pagination
	c.Context.JSON(http.StatusOK, j)
}

// Bind 将请求消息绑定到指定对象中，并做数据校验。如果校验失败，则返回校验失败错误中文文案
// 并将HTTP状态码设置成400
func (c *Context) Bind(obj interface{}) (err error) {
	return validate(c.Context.Bind(obj))
}

// ShouldBind 将请求消息绑定到指定对象中，并做数据校验。如果校验失败，则返回校验失败错误中文文案
// 类似Bind，但是不会将HTTP状态码设置成400
func (c *Context) ShouldBind(obj interface{}) (err error) {
	return validate(c.Context.ShouldBind(obj))
}


type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &defaultValidator{}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	value := reflect.ValueOf(obj)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	if valueType == reflect.Struct {
		v.lazyinit()
		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}
	return nil
}

func (v *defaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func newValidator() *validator.Validate {
	// 注册translator
	zhTranslator := zh.New()
	uni := ut.New(zhTranslator, zhTranslator)
	trans, _ = uni.GetTranslator("zh")
	validate := validator.New()
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		label := field.Tag.Get("label")
		if label == "" {
			return field.Name
		}
		return label
	})
	if err := tzh.RegisterDefaultTranslations(validate, trans); err != nil {
		elog.DefaultLogger.Fatal("Gin fail to registered Translation")
	}
	return validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = newValidator()
		v.validate.SetTagName("binding")
	})
}

var trans ut.Translator

func validate(errs error) error {
	if validationErrors, ok := errs.(validator.ValidationErrors); ok {
		var errList []string
		for _, e := range validationErrors {
			errList = append(errList, e.Translate(trans))
		}
		return errors.New(strings.Join(errList, "|"))
	}
	return errs
}

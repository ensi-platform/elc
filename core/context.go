package core

import "fmt"

type Context [][]string

func (ctx *Context) find(name string) (string, bool) {
	for _, pair := range *ctx {
		if pair[0] == name {
			return pair[1], true
		}
	}
	return "", false
}

func (ctx Context) remove(name string) Context {
	index := -1
	for i, pair := range ctx {
		if pair[0] == name {
			index = i
		}
	}

	if index > -1 {
		return append(ctx[:index], ctx[index+1:]...)
	}

	return ctx
}

func (ctx *Context) add(name string, value string) Context {
	tmp := ctx.remove(name)
	return append(tmp, []string{name, value})
}

func (ctx *Context) RenderString(str string) (string, error) {
	return substVars(str, ctx)
}

func (ctx *Context) renderMapToEnv() []string {
	var result []string
	for _, pair := range *ctx {
		result = append(result, fmt.Sprintf("%s=%s", pair[0], pair[1]))
	}

	return result
}

package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Env struct {
	loader   func(key string) (string, bool)
	problems []string
}

func New(envLoader func(key string) (string, bool)) *Env {
	return &Env{loader: envLoader}
}

func (e *Env) String(key, def string) string {
	if v, ok := e.loader(key); ok {
		return v
	}
	return def
}

func (e *Env) Err() error {
	if len(e.problems) == 0 {
		return nil
	}
	return errors.New("invalid configuration:\n  " + strings.Join(e.problems, "\n  "))
}

func (e *Env) Get(key string) (string, bool) {
	v, ok := e.loader(key)
	v = strings.TrimSpace(v)
	return v, ok
}

func (e *Env) Required(key string) string {
	v, ok := e.Get(key)
	if !ok {
		e.Invalid(key, "required, but not set")
	}
	return v
}

func (e *Env) Int(key string, def int) int {
	return Parse(e, key, def, strconv.Atoi)
}

func Parse[T any](e *Env, key string, def T, parse func(string) (T, error)) T {
	v, ok := e.Get(key)
	if !ok {
		return def
	}

	x, err := parse(v)
	if err != nil {
		e.Invalid(key, fmt.Sprintf("invalid value %q", v))
		return def
	}

	return x
}

func (e *Env) Duration(key string, def time.Duration) time.Duration {
	return Parse(e, key, def, time.ParseDuration)
}

func (e *Env) Invalid(key, problem string) {
	e.problems = append(e.problems, key+":"+problem)
}

func (e *Env) OneOf(key, def string, opts ...string) string {
	v := e.String(key, def)
	if !slices.Contains(opts, v) {
		e.Invalid(key, fmt.Sprintf("%q is not one of %s", v, strings.Join(opts, ", ")))
	}
	return def
}

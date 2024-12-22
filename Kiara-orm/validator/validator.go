package validator

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Rule representa uma regra de validação
type Rule interface {
	Validate(value interface{}) error
}

// Validator gerencia a validação de modelos
type Validator struct {
	rules map[string][]Rule
}

// NewValidator cria uma nova instância do Validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]Rule),
	}
}

// AddRule adiciona uma regra para um campo
func (v *Validator) AddRule(field string, rule Rule) {
	if v.rules[field] == nil {
		v.rules[field] = make([]Rule, 0)
	}
	v.rules[field] = append(v.rules[field], rule)
}

// Validate valida um modelo
func (v *Validator) Validate(model interface{}) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		
		// Verifica regras de validação da tag
		tag := field.Tag.Get("validate")
		if tag != "" {
			if err := v.validateField(field.Name, value.Interface(), tag); err != nil {
				return err
			}
		}
		
		// Verifica regras personalizadas
		if rules, ok := v.rules[field.Name]; ok {
			for _, rule := range rules {
				if err := rule.Validate(value.Interface()); err != nil {
					return fmt.Errorf("validação falhou para %s: %v", field.Name, err)
				}
			}
		}
	}
	
	return nil
}

// validateField valida um campo baseado na tag
func (v *Validator) validateField(field string, value interface{}, tag string) error {
	rules := strings.Split(tag, ",")
	for _, rule := range rules {
		parts := strings.Split(rule, "=")
		ruleName := parts[0]
		
		var ruleValue string
		if len(parts) > 1 {
			ruleValue = parts[1]
		}
		
		switch ruleName {
		case "required":
			if err := v.validateRequired(value); err != nil {
				return fmt.Errorf("%s: %v", field, err)
			}
		case "min":
			if err := v.validateMin(value, ruleValue); err != nil {
				return fmt.Errorf("%s: %v", field, err)
			}
		case "max":
			if err := v.validateMax(value, ruleValue); err != nil {
				return fmt.Errorf("%s: %v", field, err)
			}
		case "email":
			if err := v.validateEmail(value); err != nil {
				return fmt.Errorf("%s: %v", field, err)
			}
		}
	}
	
	return nil
}

// Implementação das regras de validação básicas
func (v *Validator) validateRequired(value interface{}) error {
	val := reflect.ValueOf(value)
	
	switch val.Kind() {
	case reflect.String:
		if val.String() == "" {
			return fmt.Errorf("campo obrigatório")
		}
	case reflect.Int, reflect.Int64:
		if val.Int() == 0 {
			return fmt.Errorf("campo obrigatório")
		}
	case reflect.Slice, reflect.Map:
		if val.Len() == 0 {
			return fmt.Errorf("campo obrigatório")
		}
	}
	
	return nil
}

func (v *Validator) validateMin(value interface{}, min string) error {
	val := reflect.ValueOf(value)
	
	switch val.Kind() {
	case reflect.String:
		minLen := len(min)
		if len(val.String()) < minLen {
			return fmt.Errorf("tamanho mínimo é %d", minLen)
		}
	case reflect.Int, reflect.Int64:
		minVal := parseInt(min)
		if val.Int() < minVal {
			return fmt.Errorf("valor mínimo é %d", minVal)
		}
	}
	
	return nil
}

func (v *Validator) validateMax(value interface{}, max string) error {
	val := reflect.ValueOf(value)
	
	switch val.Kind() {
	case reflect.String:
		maxLen := len(max)
		if len(val.String()) > maxLen {
			return fmt.Errorf("tamanho máximo é %d", maxLen)
		}
	case reflect.Int, reflect.Int64:
		maxVal := parseInt(max)
		if val.Int() > maxVal {
			return fmt.Errorf("valor máximo é %d", maxVal)
		}
	}
	
	return nil
}

func (v *Validator) validateEmail(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("valor deve ser uma string")
	}
	
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("email inválido")
	}
	
	return nil
}

// parseInt converte string para int64
func parseInt(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
} 
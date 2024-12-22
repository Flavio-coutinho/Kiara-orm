package schema

import (
	"reflect"
	"strings"
	
	"github.com/Flavio-coutinho/Kiara-orm/types"
)

// Parser é responsável por analisar as estruturas Go e extrair informações de mapeamento
type Parser struct {
	typeMapper *types.TypeMapper
}

// NewParser cria uma nova instância do Parser
func NewParser() *Parser {
	return &Parser{
		typeMapper: types.NewTypeMapper(),
	}
}

// Parse analisa uma estrutura Go e retorna seu mapeamento
func (p *Parser) Parse(model interface{}) (*types.TableMapping, error) {
	t := reflect.TypeOf(model)
	
	// Se for um ponteiro, obtém o tipo subjacente
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	// Verifica se é uma struct
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("modelo deve ser uma struct, recebido: %v", t.Kind())
	}
	
	mapping := &types.TableMapping{
		TableName: p.getTableName(t),
		Fields:    make([]types.FieldMapping, 0),
	}
	
	// Analisa cada campo da struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Ignora campos não exportados
		if !field.IsExported() {
			continue
		}
		
		fieldMapping := p.parseField(field)
		if fieldMapping != nil {
			mapping.Fields = append(mapping.Fields, *fieldMapping)
		}
	}
	
	return mapping, nil
}

// parseField analisa um campo da struct e retorna seu mapeamento
func (p *Parser) parseField(field reflect.StructField) *types.FieldMapping {
	tag := field.Tag.Get("db")
	if tag == "-" {
		return nil
	}
	
	mapping := &types.FieldMapping{
		Name: p.getFieldName(field, tag),
		Type: p.typeMapper.GetDataType(field.Type.String()),
	}
	
	// Processa as opções da tag
	p.parseTagOptions(mapping, tag)
	
	return mapping
}

// parseTagOptions processa as opções da tag db
func (p *Parser) parseTagOptions(mapping *types.FieldMapping, tag string) {
	parts := strings.Split(tag, ",")
	
	for i, part := range parts {
		if i == 0 && part != "" {
			mapping.Name = part
			continue
		}
		
		switch {
		case part == "primarykey":
			mapping.IsPrimaryKey = true
		case part == "autoincrement":
			mapping.IsAutoInc = true
		case part == "unique":
			mapping.IsUnique = true
		case part == "nullable":
			mapping.IsNullable = true
		case strings.HasPrefix(part, "size:"):
			size, _ := strconv.Atoi(strings.TrimPrefix(part, "size:"))
			mapping.Size = size
		}
	}
}

// getTableName retorna o nome da tabela para a struct
func (p *Parser) getTableName(t reflect.Type) string {
	// Primeiro tenta encontrar uma tag de tabela na struct
	if tableTag, ok := t.FieldByName("TableName"); ok {
		if tag := tableTag.Tag.Get("db"); tag != "" {
			return tag
		}
	}
	
	// Caso contrário, usa o nome da struct em minúsculo
	return strings.ToLower(t.Name())
}

// getFieldName retorna o nome do campo para o banco de dados
func (p *Parser) getFieldName(field reflect.StructField, tag string) string {
	if tag != "" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}
	return strings.ToLower(field.Name)
}

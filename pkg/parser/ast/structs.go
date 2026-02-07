package ast

type Attribute struct {
	Name       string
	Type       TypeRef
	Visibility Visibility
}

type Method struct {
	Name       string
	ReturnType TypeRef
	Parameters []Attribute
	Visibility Visibility
}

type Class struct {
	Name       string
	Attributes []Attribute
	Methods    []Method
}

func (c *Class) String() string {
	return c.Name
}

type Relationship struct {
	From *Class
	To   *Class
	Type RelationType

	MultiplicityFrom Multiplicity
	MultiplicityTo   Multiplicity

	Comment string
}

type RelationshipsRepo map[string][]*Relationship

func (r RelationshipsRepo) AddRelationship(rel *Relationship) {
	r[rel.From.Name] = append(r[rel.From.Name], rel)
	r[rel.To.Name] = append(r[rel.To.Name], rel)
}

func (r RelationshipsRepo) GetRelationships(className string) ([]*Relationship, bool) {
	rel, exists := r[className]
	return rel, exists
}

type ClassDiagram struct {
	Classes       map[string]*Class
	Relationships RelationshipsRepo
}

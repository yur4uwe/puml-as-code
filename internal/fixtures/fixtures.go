package fixtures

// ComprehensivePUML is a representative PlantUML diagram containing comments,
// preprocessor/render commands, packages, entities (interfaces, classes, enums),
// stereotypes, styles, members, notes, and diverse relationships with multiplicities.
const ComprehensivePUML = `@startuml Comprehensive Diagram
' Single line diagram header comment
/' Multi-line block comment
   explaining the architecture
   and domain design '/

top to bottom direction
hide empty members
scale 1.5
skinparam classAttributeIconSize 0

package service {
  ' Greeter interface definition
  interface Greeter <<service>> {
    +Greet(name string) string
  }

  /' UserService handles user management
     and integrates with Repository '/
  class UserService <<service>> #aliceblue {
    +Repo Repository
    +Timeout int
    +Greet(name string) string
    +FindUser(id string) (*User, error)
  }

  class Repository <<repository>> {
    +DB string
    +Query(id string) (*User, error)
  }

  class User <<entity>> {
    +ID string
    +Name string
    +Active bool
    +Status Status
  }

  enum Status {
    +Active
    +Inactive
  }
}

note "Service layer architecture" as N1
note right of UserService: Core domain logic

UserService ..|> Greeter : implements
UserService *-- "1" Repository : owns
UserService ..> User : returns
Repository "1" --> "0..*" User : queries
@enduml
`

# Class Diagram - Relations Between Classes

## Relationship Types

Relations between classes are defined using the following symbols:

| Type           | Symbol   | Purpose                                       |
| -------------- | -------- | --------------------------------------------- |
| Extension      | ``<\|--`` | Specialization of a class in a hierarchy      |
| Implementation | ``<\|..`` | Realization of an interface by a class        |
| Composition    | ``*--``  | The part cannot exist without the whole       |
| Aggregation    | ``o--``  | The part can exist independently of the whole |
| Dependency     | ``-->``  | The object uses another object                |
| Dependency     | ``..>``  | A weaker form of dependency                   |

It is possible to replace ``--`` by ``..`` to have a dotted line.

## Basic Relationships

<plantuml>
@startuml
Class01 <|-- Class02
Class03 *-- Class04
Class05 o-- Class06
Class07 .. Class08
Class09 -- Class10
@enduml
</plantuml>

<plantuml>
@startuml
Class11 <|.. Class12
Class13 --> Class14
Class15 ..> Class16
Class17 ..|> Class18
Class19 <--* Class20
@enduml
</plantuml>

## Decorated Relations

PlantUML supports decorated relationship symbols for directional hints:

<plantuml>
@startuml
Class21 #-- Class22
Class23 x-- Class24
Class25 }-- Class26
Class27 +-- Class28
Class29 ^-- Class30
@enduml
</plantuml>

These decorated relations allow for more expressive relationship semantics while maintaining visual clarity.

## Labels on Relations

It is possible to add a label on the relation, using ``:``, followed by the text of the label.

For cardinality, you can use double-quotes ``""`` on each side of the relation.

<plantuml>
@startuml

Class01 "1" *-- "many" Class02 : contains

Class03 o-- Class04 : aggregation

Class05 --> "1" Class06

@enduml
</plantuml>

You can add an extra arrow pointing at one object showing which object acts on the other object, using ``<`` or ``>`` at the begin or at the end of the label.

<plantuml>
@startuml
class Car

Driver - Car : drives >
Car *- Wheel : have 4 >
Car -- Person : < owns

@enduml
</plantuml>

## Changing Arrows Orientation

By default, links between classes have two dashes ``--`` and are vertically oriented. It is possible to use horizontal link by putting a single dash (or dot) like this:

<plantuml>
@startuml
Room o- Student
Room *-- Chair
@enduml
</plantuml>

You can also change directions by reversing the link:

<plantuml>
@startuml
Student -o Room
Chair --* Room
@enduml
</plantuml>

It is also possible to change arrow direction by adding ``left``, ``right``, ``up`` or ``down`` keywords inside the arrow:

<plantuml>
@startuml
foo -left-> dummyLeft
foo -right-> dummyRight
foo -up-> dummyUp
foo -down-> dummyDown
@enduml
</plantuml>

You can shorten the arrow by using only the first character of the direction (for example, ``-d-`` instead of ``-down-``) or the two first characters (``-do-``).

Please note that you should not abuse this functionality : *Graphviz* gives usually good results without tweaking.

And with the [``left to right direction``](use-case-diagram#d551e48d272b2b07) parameter: 

<plantuml>
@startuml
left to right direction
foo -left-> dummyLeft
foo -right-> dummyRight
foo -up-> dummyUp
foo -down-> dummyDown
@enduml
</plantuml>

## Association Classes

You can define *association class* after that a relation has been defined between two classes, like in this example:

<plantuml>
@startuml
class Student {
  Name
}
Student "0..*" - "1..*" Course
(Student, Course) .. Enrollment

class Enrollment {
  drop()
  cancel()
}
@enduml
</plantuml>

You can define it in another direction:

<plantuml>
@startuml
class Student {
  Name
}
Student "0..*" -- "1..*" Course
(Student, Course) . Enrollment

class Enrollment {
  drop()
  cancel()
}
@enduml
</plantuml>

## Association on Same Class

<plantuml>
@startuml
class Station {
    +name: string
}

class StationCrossing {
    +cost: TimeInterval
}

<> diamond

StationCrossing . diamond
diamond - "from 0..*" Station
diamond - "to 0..* " Station
@enduml
</plantuml>

*[Ref. [Incubation: Associations](http://wiki.plantuml.net/site/incubation#associations)]*

## Lollipop Interface

You can also define lollipops interface on classes, using the following syntax:
* ``bar ()- foo``
* ``bar ()-- foo``
* ``foo -() bar``

<plantuml>
@startuml
class foo
bar ()- foo
@enduml
</plantuml>

## Qualified Associations

### Minimal Example

<plantuml>
@startuml
class class1
class class2

class1 [Qualifier] - class2
@enduml
</plantuml>

*[Ref. [QA-16397](https://forum.plantuml.net/16397/add-qualified-associations-to-class-diagrams), [GH-1467](https://github.com/plantuml/plantuml/issues/1467)]*

### Another Example

<plantuml>
@startuml
    interface Map<K,V>
    class HashMap<Long,Customer>

    Map <|.. HashMap
    Shop [customerId: long] ---> "customer\n1" Customer
    HashMap [id: Long] -r-> "value" Customer
@enduml
</plantuml>

## Arrows from/to Class Members

<plantuml>
@startuml
class Foo {
+ field1
+ field2
}

class Bar {
+ field3
+ field4
}

Foo::field1 --> Bar::field3 : foo
Foo::field2 --> Bar::field4 : bar
@enduml
</plantuml>

*[Ref. [QA-3636](https://forum.plantuml.net/3636)]*

<plantuml>
@startuml
left to right direction

class User {
  id : INTEGER
  ..
  other_id : INTEGER
}

class Email {
  id : INTEGER
  ..
  user_id : INTEGER
  address : INTEGER
}

User::id *-- Email::user_id
@enduml
</plantuml>

*[Ref. [QA-5261](https://forum.plantuml.net/5261)]*

## Grouping Inheritance Arrow Heads

You can merge all arrow heads using the `skinparam groupInheritance`, with a threshold as parameter.

### GroupInheritance 1 (no grouping)
<plantuml>
@startuml
skinparam groupInheritance 1

A1 <|-- B1

A2 <|-- B2
A2 <|-- C2

A3 <|-- B3
A3 <|-- C3
A3 <|-- D3

A4 <|-- B4
A4 <|-- C4
A4 <|-- D4
A4 <|-- E4
@enduml
</plantuml>

### GroupInheritance 2 (grouping from 2)
<plantuml>
@startuml
skinparam groupInheritance 2

A1 <|-- B1

A2 <|-- B2
A2 <|-- C2

A3 <|-- B3
A3 <|-- C3
A3 <|-- D3

A4 <|-- B4
A4 <|-- C4
A4 <|-- D4
A4 <|-- E4
@enduml
</plantuml>

### GroupInheritance 3 (grouping only from 3)
<plantuml>
@startuml
skinparam groupInheritance 3

A1 <|-- B1

A2 <|-- B2
A2 <|-- C2

A3 <|-- B3
A3 <|-- C3
A3 <|-- D3

A4 <|-- B4
A4 <|-- C4
A4 <|-- D4
A4 <|-- E4
@enduml
</plantuml>

### GroupInheritance 4 (grouping only from 4)
<plantuml>
@startuml
skinparam groupInheritance 4

A1 <|-- B1

A2 <|-- B2
A2 <|-- C2

A3 <|-- B3
A3 <|-- C3
A3 <|-- D3

A4 <|-- B4
A4 <|-- C4
A4 <|-- D4
A4 <|-- E4
@enduml
</plantuml>

*[Ref. [QA-3193](https://forum.plantuml.net/3193/grouping-inheritance-arrow-ends), and Defect [QA-13532](https://forum.plantuml.net/13532/groupinheritance-bug)]*

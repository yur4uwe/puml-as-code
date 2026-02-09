# Class Diagram - Notes and Stereotypes

## Basic Notes and Stereotypes

Stereotypes are defined with the ``class`` keyword, ``<<`` and ``>>``.

You can also define notes using ``note left of``, ``note right of``, ``note top of``, ``note bottom of`` keywords.

You can also define a note on the last defined element using ``note left``, ``note right``, ``note top``, ``note bottom``.

A note can be also define alone with the ``note`` keywords, then linked to other objects using the ``..`` symbol.

<plantuml>
@startuml
class Object << general >>
Object <|--- ArrayList

note top of Object : In java, every class\nextends this one.

note "This is a floating note" as N1
note "This note is connected\nto several objects." as N2
Object .. N2
N2 .. ArrayList

class Foo
note left: On last defined class

@enduml
</plantuml>

## More on Notes

It is also possible to use few HTML tags (See [Creole expression](creole)) like:

* ``<b>``
* ``<u>``
* ``<i>``
* ``<s>``, ``<del>``, ``<strike>``
* ``<font color="#AAAAAA">`` or ``<font color="colorName">``
* ``<color:#AAAAAA>`` or ``<color:colorName>``
* ``<size:nn>`` to change font size
* ``<img src="file">`` or ``<img:file>``: the file must be accessible by the filesystem

You can also have a note on several lines.

You can also define a note on the last defined element using ``note left``, ``note right``, ``note top``, ``note bottom``.

<plantuml>
@startuml

class Foo
note left: On last defined class

note top of Foo
  In java, <size:18>every</size> <u>class</u>
  <b>extends</b>
  <i>this</i> one.
end note

note as N1
  This note is <u>also</u>
  <b><color:royalBlue>on several</color>
  <s>words</s> lines
  And this is hosted by <img:https://plantuml.com/sourceforge.jpg>
end note

@enduml
</plantuml>

## Note on Field or Method

It is possible to add a note on field (field, attribute, member) or on method.

### ⚠ Constraint
* This cannot be used with ``top`` or ``bottom`` *(only ``left`` and ``right`` are implemented)*
* This cannot be used with namespaceSeparator ``::``

### Note on Field or Method

<plantuml>
@startuml
class A {
{static} int counter
+void {abstract} start(int timeout)
}
note right of A::counter
  This member is annotated
end note
note right of A::start
  This method is now explained in a UML note
end note
@enduml
</plantuml>

### Note on Method with the Same Name

<plantuml>
@startuml
class A {
{static} int counter
+void {abstract} start(int timeoutms)
+void {abstract} start(Duration timeout)
}
note left of A::counter
  This member is annotated
end note
note right of A::"start(int timeoutms)"
  This method with int
end note
note right of A::"start(Duration timeout)"
  This method with Duration
end note
@enduml
</plantuml>

*[Ref. [QA-3474](https://forum.plantuml.net/3474) and [QA-5835](https://forum.plantuml.net/5835)]*

## Note on Links

It is possible to add a note on a link, just after the link definition, using ``note on link``.

You can also use ``note left on link``, ``note right on link``, ``note top on link``, ``note bottom on link`` if you want to change the relative position of the note with the label.

<plantuml>
@startuml

class Dummy
Dummy --> Foo : A link
note on link #red: note that is red

Dummy --> Foo2 : Another link
note right on link #blue
this is my note on right link
and in blue
end note

@enduml
</plantuml>

## Abstract Class and Interface

You can declare a class as abstract using ``abstract`` or ``abstract class`` keywords.

The class will be printed in *italic*.

You can use the ``interface``, ``annotation`` and ``enum`` keywords too.

<plantuml>
@startuml

abstract class AbstractList
abstract AbstractCollection
interface List
interface Collection

List <|-- AbstractList
Collection <|-- AbstractCollection

Collection <|- List
AbstractCollection <|- AbstractList
AbstractList <|-- ArrayList

class ArrayList {
  Object[] elementData
  size()
}

enum TimeUnit {
  DAYS
  HOURS
  MINUTES
}

annotation SuppressWarnings

annotation Annotation {
  annotation with members
  String foo()
  String bar()
}

@enduml
</plantuml>

*[Ref. 'Annotation with members' [Issue#458](https://github.com/plantuml/plantuml/issues/458)]*

# Class Diagram - Hide, Show, and Remove

## Hide/Show Commands

You can parameterize the display of classes using the ``hide/show`` command.

The basic command is: ``hide empty members``. This command will hide attributes or methods if they are empty.

Instead of ``empty members``, you can use:
* ``empty fields`` or ``empty attributes`` for empty fields,
* ``empty methods`` for empty methods,
* ``fields`` or ``attributes`` which will hide fields, even if they are described,
* ``methods`` which will hide methods, even if they are described,
* ``members`` which will hide fields __and__ methods, even if they are described,
* ``circle`` for the circled character in front of class name,
* ``stereotype`` for the stereotypee

You can also provide, just after the ``hide`` or ``show`` keyword:
* ``class`` for all classes,
* ``interface`` for all interfaces,
* ``enum`` for all enums,
* ``<<foo1>>`` for classes which are stereotyped with *foo1*,
* an existing class name.

You can use several ``show/hide`` commands to define rules and exceptions.

<plantuml>
@startuml

class Dummy1 {
  +myMethods()
}

class Dummy2 {
  +hiddenMethod()
}

class Dummy3 <<Serializable>> {
String name
}

hide members
hide <<Serializable>> circle
show Dummy1 methods
show <<Serializable>> fields

@enduml
</plantuml>

You can also mix with visibility:

<plantuml>
@startuml
hide private members
hide protected members
hide package members

class Foo {
  - private
  # protected
  ~ package
}
@enduml
</plantuml>

*[Ref. [QA-2913](https://forum.plantuml.net/2913/hiding-based-on-visibilty?show=2916#a2916)]*

## Hide Classes

You can also use the ``show/hide`` commands to hide classes.

This may be useful if you define a large [!included file](preprocessing), and if you want to hide some classes after [file inclusion](preprocessing).

<plantuml>
@startuml

class Foo1
class Foo2

Foo2 *-- Foo1

hide Foo2

@enduml
</plantuml>

## Remove Classes

You can also use the ``remove`` commands to remove classes.

This may be useful if you define a large [!included file](preprocessing), and if you want to remove some classes after [file inclusion](preprocessing).

<plantuml>
@startuml

class Foo1
class Foo2

Foo2 *-- Foo1

remove Foo2

@enduml
</plantuml>

## Hide, Remove or Restore Tagged Element or Wildcard

You can put `$tags` (using `$`) on elements, then remove, hide or restore components either individually or by tags.

By default, all components are displayed:

<plantuml>
@startuml
class C1 $tag13
enum E1
interface I1 $tag13
C1 -- I1
@enduml
</plantuml>

But you can:
* `hide $tag13` components:

<plantuml>
@startuml
class C1 $tag13
enum E1
interface I1 $tag13
C1 -- I1

hide $tag13
@enduml
</plantuml>

* or `remove $tag13` components:

<plantuml>
@startuml
class C1 $tag13
enum E1
interface I1 $tag13
C1 -- I1

remove $tag13
@enduml
</plantuml>

* or `remove $tag13 and restore $tag1` components:

<plantuml>
@startuml
class C1 $tag13 $tag1
enum E1
interface I1 $tag13
C1 -- I1

remove $tag13
restore $tag1
@enduml
</plantuml>

* or ``remove * and restore $tag1`` components:

<plantuml>
@startuml
class C1 $tag13 $tag1
enum E1
interface I1 $tag13
C1 -- I1

remove *
restore $tag1
@enduml
</plantuml>

## Hide or Remove Unlinked Class

By default, all classes are displayed:

<plantuml>
@startuml
class C1
class C2
class C3
C1 -- C2
@enduml
</plantuml>

But you can:
* `hide @unlinked` classes:

<plantuml>
@startuml
class C1
class C2
class C3
C1 -- C2

hide @unlinked
@enduml
</plantuml>

* or `remove @unlinked` classes:

<plantuml>
@startuml
class C1
class C2
class C3
C1 -- C2

remove @unlinked
@enduml
</plantuml>

*[Adapted from [QA-11052](https://forum.plantuml.net/11052)]*

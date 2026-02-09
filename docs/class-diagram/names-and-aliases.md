# Class Diagram - Names and Aliases

## Using Non-Letters in Element Names and Relation Labels

If you want to use [non-letters](unicode) in the class (or enum...) display name, you can either:
* Use the ``as`` keyword in the class definition to assign an alias
* Put quotes ``""`` around the class name

<plantuml>
@startuml
class "This is my class" as class1
class class2 as "It works this way too"

class2 *-- "foo/dummy" : use
@enduml
</plantuml>

If an alias is assigned to an element, the rest of the file must refer to the element by the alias instead of the name.

## Starting Names with ``$``

Note that names starting with ``$`` cannot be hidden or removed later, because ``hide`` and ``remove`` command will consider the name a ``$tag`` instead of a component name. To later remove such elements they must have an alias or must be tagged.

<plantuml>
@startuml
class $C1
class $C2 $C2
class "$C2" as dollarC2
remove $C1
remove $C2
remove dollarC2
@enduml
</plantuml>

Also note that names starting with ``$`` are valid, but to assign an alias to such element the name must be put between quotes ``""``.

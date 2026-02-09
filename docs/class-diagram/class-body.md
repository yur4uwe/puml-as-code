# Class Diagram - Class Body

## Adding Methods and Fields

To declare fields and methods, you can use the symbol ``:`` followed by the field's or method's name.

The system checks for parenthesis to choose between methods and fields.

<plantuml>
@startuml
Object <|-- ArrayList

Object : equals()
ArrayList : Object[] elementData
ArrayList : size()

@enduml
</plantuml>

It is also possible to group between brackets ``{}`` all fields and methods.

Note that the syntax is highly flexible about type/name order.

<plantuml>
@startuml
class Dummy {
  String data
  void methods()
}

class Flight {
   flightNumber : Integer
   departureTime : Date
}
@enduml
</plantuml>

You can use ``{field}`` and ``{method}`` modifiers to override default behaviour of the parser about fields and methods.

<plantuml>
@startuml
class Dummy {
  {field} A field (despite parentheses)
  {method} Some method
}

@enduml
</plantuml>

## Defining Visibility

### Visibility for Methods or Fields

When you define methods or fields, you can use characters to define the visibility of the corresponding item:

| Character | Icon for field                 | Icon for method                 | Visibility          |
| --------- | ------------------------------ | ------------------------------- | ------------------- |
| ``-``     | ![](private-field.png)         | ![](private-method.png)         | ``private``         |
| ``#``     | ![](protected-field.png)       | ![](protected-method.png)       | ``protected``       |
| ``~``     | ![](package-private-field.png) | ![](package-private-method.png) | ``package private`` |
| ``+``     | ![](public-field.png)          | ![](public-method.png)          | ``public``          |

<plantuml>
@startuml

class Dummy {
 -field1
 #field2
 ~method1()
 +method2()
}

@enduml
</plantuml>

You can turn off this feature using the ``skinparam classAttributeIconSize 0`` command:

<plantuml>
@startuml
skinparam classAttributeIconSize 0
class Dummy {
 -field1
 #field2
 ~method1()
 +method2()
}

@enduml
</plantuml>

Visibility indicators are optional and can be ommitted individualy without turning off the icons globally using `skinparam classAttributeIconSize 0`.

<plantuml>
@startuml
class Dummy {
 field1
 field2
 method1()
 method2()
}

@enduml
</plantuml>

In such case if you'd like to use methods or fields that start with `-`, `#`, `~` or `+` characters such as a destructor in some languages for `Dummy` class `~Dummy()`, escape the first character with a `\\` character:

<plantuml>
@startuml
class Dummy {
 field1
 \~Dummy()
 method1()
}

@enduml
</plantuml>

### Visibility for Class

Similar to methods or fields, you can use same characters to define the Class visibility:

<plantuml>
@startuml
-class "private Class" {
}

#class "protected Class" {
}

~class "package private Class" {
}

+class "public Class" {
}
@enduml
</plantuml>

*[Ref. [QA-4755](https://forum.plantuml.net/4755/provide-display-visibility-attributes-private-protected)]*

## Abstract and Static

You can define static or abstract methods or fields using the ``{static}`` or ``{abstract}`` modifier.

These modifiers can be used at the start or at the end of the line. You can also use ``{classifier}`` instead of ``{static}``.

<plantuml>
@startuml
class Dummy {
  {static} String id
  {abstract} void methods()
}
@enduml
</plantuml>

## Advanced Class Body

By default, methods and fields are automatically regrouped by PlantUML. You can use separators to define your own way of ordering fields and methods. The following separators are possible: ``--`` ``..`` ``==`` ``__``.

You can also use titles within the separators:

<plantuml>
@startuml
class Foo1 {
  You can use
  several lines
  ..
  as you want
  and group
  ==
  things together.
  __
  You can have as many groups
  as you want
  --
  End of class
}

class User {
  .. Simple Getter ..
  + getName()
  + getAddress()
  .. Some setter ..
  + setName()
  __ private data __
  int age
  -- encrypted --
  String password
}

@enduml
</plantuml>

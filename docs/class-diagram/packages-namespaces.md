# Class Diagram - Packages and Namespaces

## Packages

You can define a package using the ``package`` keyword, and optionally declare a background color for your package (Using a html color code or name).

Note that package definitions can be nested.

<plantuml>
@startuml

package "Classic Collections" #DDDDDD {
  Object <|-- ArrayList
}

package com.plantuml {
  Object <|-- Demo1
  Demo1 *- Demo2
}

@enduml
</plantuml>

## Packages Style

There are different styles available for packages.

You can specify them either by setting a default style with the command: ``skinparam packageStyle``, or by using a stereotype on the package:

<plantuml>
@startuml
scale 750 width
package foo1 <<Node>> {
  class Class1
}

package foo2 <<Rectangle>> {
  class Class2
}

package foo3 <<Folder>> {
  class Class3
}

package foo4 <<Frame>> {
  class Class4
}

package foo5 <<Cloud>> {
  class Class5
}

package foo6 <<Database>> {
  class Class6
}

@enduml
</plantuml>

You can also define links between packages, like in the following example:

<plantuml>
@startuml

skinparam packageStyle rectangle

package foo1.foo2 {
}

package foo1.foo2.foo3 {
  class Object
}

foo1.foo2 +-- foo1.foo2.foo3

@enduml
</plantuml>

## Namespaces

Starting with version 1.2023.2 (which is online as a beta), PlantUML handles differently namespaces and packages.

There won't be any difference between namespaces and packages anymore: both keywords are now synonymous.

## Automatic Package Creation

You can define another separator (other than the dot) using the command: ``set separator ???``.

<plantuml>
@startuml

set separator ::
class X1::X2::foo {
  some info
}

@enduml
</plantuml>

You can disable automatic namespace creation using the command ``set separator none``.

<plantuml>
@startuml

set separator none
class X1.X2.foo {
  some info
}

@enduml
</plantuml>

## Packages and Namespaces Enhancement

*[From V1.2023.2+, and V1.2023.5]*

<plantuml>
@startuml
class A.B.C.D.Z {
}
@enduml
</plantuml>

<plantuml>
@startuml
set separator none
class A.B.C.D.Z {
}
@enduml
</plantuml>

<plantuml>
@startuml
!pragma useIntermediatePackages false
class A.B.C.D.Z {
}
@enduml
</plantuml>

<plantuml>
@startuml
set separator none
package A.B.C.D {
  class Z {
  }
}
@enduml
</plantuml>

*[Ref. [GH-1352](https://github.com/plantuml/plantuml/issues/1352)]*

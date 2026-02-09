# Class Diagram - Layout and Orientation

## Help on Layout

Sometimes, the default layout is not perfect...

You can use ``together`` keyword to group some classes together: the layout engine will try to group them (as if they were in the same package).

You can also use ``hidden`` links to force the layout.

<plantuml>
@startuml

class Bar1
class Bar2
together {
  class Together1
  class Together2
  class Together3
}
Together1 - Together2
Together2 - Together3
Together2 -[hidden]--> Bar1
Bar1 -[hidden]> Bar2

@enduml
</plantuml>

## Change Diagram Orientation

You can change (whole) diagram orientation with: 
- `top to bottom direction` _(by default)_
- `left to right direction`

### Top to Bottom _(by default)_

#### With [Graphviz](graphviz-dot) _(layout engine by default)_

The main rule is: **Nested element first, then simple element.**

<plantuml>
@startuml
class a
class b
package A {
  class a1
  class a2
  class a3
  class a4
  class a5
  package sub_a {
   class sa1
   class sa2
   class sa3
  }
}
  
package B {
  class b1
  class b2
  class b3
  class b4
  class b5
  package sub_b {
   class sb1
   class sb2
   class sb3
  }
}
@enduml
</plantuml>

#### With [Smetana](smetana02) _(internal layout engine)_

The main rule is the opposite: **Simple element first, then nested element.**

<plantuml>
@startuml
!pragma layout smetana
class a
class b
package A {
  class a1
  class a2
  class a3
  class a4
  class a5
  package sub_a {
   class sa1
   class sa2
   class sa3
  }
}
  
package B {
  class b1
  class b2
  class b3
  class b4
  class b5
  package sub_b {
   class sb1
   class sb2
   class sb3
  }
}
@enduml
</plantuml>

### Left to Right

#### With [Graphviz](graphviz-dot) _(layout engine by default)_

<plantuml>
@startuml
left to right direction
class a
class b
package A {
  class a1
  class a2
  class a3
  class a4
  class a5
  package sub_a {
   class sa1
   class sa2
   class sa3
  }
}
  
package B {
  class b1
  class b2
  class b3
  class b4
  class b5
  package sub_b {
   class sb1
   class sb2
   class sb3
  }
}
@enduml
</plantuml>

#### With [Smetana](smetana02) _(internal layout engine)_

<plantuml>
@startuml
!pragma layout smetana
left to right direction
class a
class b
package A {
  class a1
  class a2
  class a3
  class a4
  class a5
  package sub_a {
   class sa1
   class sa2
   class sa3
  }
}
  
package B {
  class b1
  class b2
  class b3
  class b4
  class b5
  package sub_b {
   class sb1
   class sb2
   class sb3
  }
}
@enduml
</plantuml>

# Class Diagram - Styling

## Skinparam

You can use the [skinparam](skinparam) command to change colors and fonts for the drawing.

You can use this command:
* In the diagram definition, like any other commands,
* In an [included file](preprocessing),
* In a configuration file, provided in the [command line](command-line) or the [ANT task](ant-task).

<plantuml>
@startuml

skinparam class {
BackgroundColor PaleGreen
ArrowColor SeaGreen
BorderColor SpringGreen
}
skinparam stereotypeCBackgroundColor YellowGreen

Class01 "1" *-- "many" Class02 : contains

Class03 o-- Class04 : aggregation

@enduml
</plantuml>

## Skinned Stereotypes

You can define specific color and fonts for stereotyped classes.

<plantuml>
@startuml

skinparam class {
BackgroundColor PaleGreen
ArrowColor SeaGreen
BorderColor SpringGreen
BackgroundColor<<Foo>> Wheat
BorderColor<<Foo>> Tomato
}
skinparam stereotypeCBackgroundColor YellowGreen
skinparam stereotypeCBackgroundColor<< Foo >> DimGray

class Class01 <<Foo>>
class Class03 <<Foo>>
Class01 "1" *-- "many" Class02 : contains

Class03 o-- Class04 : aggregation

@enduml
</plantuml>

**Important**: unlike class stereotypes, there must be no space between the skin parameter and the following stereotype. 

Any of the spaces shown as `_` below will cause **all** skinparams to be ignored, 
see [discord discussion](https://discord.com/channels/1083727021328306236/1289954399321329755/1289967399302467614)
and [issue #1932](https://github.com/plantuml/plantuml/issues/1932):
- `BackgroundColor_<<Foo>> Wheat`
- `skinparam stereotypeCBackgroundColor_<<Foo>> DimGray`

## Color Gradient

You can declare individual colors for classes, notes etc using the \# notation.

You can use standard color names or RGB codes in various notations, see [Colors](color).

You can also use color gradient for background colors, with the following syntax:
two colors names separated either by:
* ``|``,
* ``/``,
* ``\``, or 
* ``-``
depending on the direction of the gradient.

For example:

<plantuml>
@startuml

skinparam backgroundcolor AntiqueWhite/Gold
skinparam classBackgroundColor Wheat|CornflowerBlue

class Foo #red-green
note left of Foo #blue\9932CC
  this is my
  note on this class
end note

package example #GreenYellow/LightGoldenRodYellow {
  class Dummy
}

@enduml
</plantuml>

## Bracketed Relations (Linking or Arrow) Style

### Line Style

It's also possible to have explicitly `bold`, `dashed`, `dotted`, `hidden` or `plain` relation, links or arrows:  

* without label

<plantuml>
@startuml
title Bracketed line style without label
class foo
class bar
bar1 : [bold]  
bar2 : [dashed]
bar3 : [dotted]
bar4 : [hidden]
bar5 : [plain] 

foo --> bar
foo -[bold]-> bar1
foo -[dashed]-> bar2
foo -[dotted]-> bar3
foo -[hidden]-> bar4
foo -[plain]-> bar5
@enduml
</plantuml>

* with label

<plantuml>
@startuml
title Bracketed line style with label
class foo
class bar
bar1 : [bold]  
bar2 : [dashed]
bar3 : [dotted]
bar4 : [hidden]
bar5 : [plain] 

foo --> bar          : ∅
foo -[bold]-> bar1   : [bold]
foo -[dashed]-> bar2 : [dashed]
foo -[dotted]-> bar3 : [dotted]
foo -[hidden]-> bar4 : [hidden]
foo -[plain]-> bar5  : [plain]

@enduml
</plantuml>

*[Adapted from [QA-4181](https://forum.plantuml.net/4181/how-change-width-line-in-a-relationship-between-two-classes?show=4232#a4232)]*

### Line Color

<plantuml>
@startuml
title Bracketed line color
class foo
class bar
bar1 : [#red]
bar2 : [#green]
bar3 : [#blue]

foo --> bar
foo -[#red]-> bar1     : [#red]
foo -[#green]-> bar2   : [#green]
foo -[#blue]-> bar3    : [#blue]
@enduml
</plantuml>

### Line Thickness

<plantuml>
@startuml
title Bracketed line thickness
class foo
class bar
bar1 : [thickness=1]
bar2 : [thickness=2]
bar3 : [thickness=4]
bar4 : [thickness=8]
bar5 : [thickness=16]

foo --> bar                 : ∅
foo -[thickness=1]-> bar1   : [1]
foo -[thickness=2]-> bar2   : [2]
foo -[thickness=4]-> bar3   : [4]
foo -[thickness=8]-> bar4   : [8]
foo -[thickness=16]-> bar5  : [16]

@enduml
</plantuml>

*[Ref. [QA-4949](https://forum.plantuml.net/4949)]*

### Mix

<plantuml>
@startuml
title Bracketed line style mix
class foo
class bar
bar1 : [#red,thickness=1]
bar2 : [#red,dashed,thickness=2]
bar3 : [#green,dashed,thickness=4]
bar4 : [#blue,dotted,thickness=8]
bar5 : [#blue,plain,thickness=16]

foo --> bar                             : ∅
foo -[#red,thickness=1]-> bar1          : [#red,1]
foo -[#red,dashed,thickness=2]-> bar2   : [#red,dashed,2]
foo -[#green,dashed,thickness=4]-> bar3 : [#green,dashed,4]
foo -[#blue,dotted,thickness=8]-> bar4  : [blue,dotted,8]
foo -[#blue,plain,thickness=16]-> bar5  : [blue,plain,16]
@enduml
</plantuml>

## Change Relation (Linking or Arrow) Color and Style (Inline Style)

You can change the [color](color) or style of individual relation or arrows using the inline following notation:

* `#color;line.[bold|dashed|dotted];text:color`

<plantuml>
@startuml
class foo
foo --> bar : normal
foo --> bar1 #line:red;line.bold;text:red  : red bold
foo --> bar2 #green;line.dashed;text:green : green dashed
foo --> bar3 #blue;line.dotted;text:blue   : blue dotted
@enduml
</plantuml>

*[See similar feature on [deployment](deployment-diagram#0b2e57c3d4eafdda)]*

## Change Class Color and Style (Inline Style)

You can change the [color](color) or style of individual class using the two following notations: 

* `#color ##[style]color` 

With background color first (`#color`), then line style and line color (`##[style]color` )

<plantuml>
@startuml
abstract   abstract
annotation annotation #pink ##[bold]red
class      class      #palegreen ##[dashed]green
interface  interface  #aliceblue ##[dotted]blue
@enduml
</plantuml>

*[Ref. [QA-1487](https://forum.plantuml.net/1487)]*

* `#[color|back:color];header:color;line:color;line.[bold|dashed|dotted];text:color`

<plantuml>
@startuml
abstract   abstract
annotation annotation #pink;line:red;line.bold;text:red
class      class      #palegreen;line:green;line.dashed;text:green
interface  interface  #aliceblue;line:blue;line.dotted;text:blue
@enduml
</plantuml>

First original example:

<plantuml>
@startuml
class bar #line:green;back:lightblue
class bar2 #lightblue;line:green

class Foo1 #back:red;line:00FFFF
class FooDashed #line.dashed:blue
class FooDotted #line.dotted:blue
class FooBold #line.bold
class Demo1 #back:lightgreen|yellow;header:blue/red
@enduml
</plantuml>

*[Ref. [QA-3770](https://forum.plantuml.net/3770)]*

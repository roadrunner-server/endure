Happy scenario graph

Dependency graph of the structs in the folder  

digraph G {  
  "foo1.S1" -> "foo4.S4"  
  "foo1.S1" -> "foo2.S2"  
  "foo2.S2" -> "foo4.S4"  
  "foo3.S2" -> "foo4.S4"  
  "foo3.S2" -> "foo2.S2"  
  "foo4.S2" -> "foo5.S5"  
  "foo6.S6"
}

<p align="left">
  <img src="https://github.com/spiral/endure/blob/master/images/happyPathGraph.png" width="300" height="250" />
</p>

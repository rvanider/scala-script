#!/usr/bin/env scala-script
//#include included-top.scala
object FromHere {
  override def toString = "from-here"
}

println(FromHere)
println(FromTop)
println(FromOne)
println(FromTwo)
println(FromDeeper)
println(FolderIncluded)
println(FolderSourced)
println(FolderContent)

import com.example.two.ExampleTwo
import com.example.one.ExampleOne

println(ExampleOne)
println(ExampleTwo)

args foreach println

# scala-script

Scala-script is a standalone binary that can be used to launch
[Scala](http://www.scala-lang.org) scripts with support for automatic
class path management and a simplified script include mechanism.

## Dependencies

The `scala-script` binary has no dependencies other than scala being
installed and on the path.

- Linux and OSX have been tested
- Scala 2.11.8 has been tested

## Installation

The only file in this repository you need is the `scala-script` for
your respective platform (OSX or Linux).

    curl -s --fail -L --output "/path/to/store/scala-script" https://github.com/rvanider/scala-script/blob/master/scala-script?raw=true
    chmod +x "/path/to/store/scala-script"

## Features

- `classpath` - the classpath will automatically be built for any jar files located in the `lib/` folder
  relative to the script being executed
- `include` - other scala scripts can be included using the syntax `//#include path/to/file.scala`
- `repl` - launching scala-script with `--repl` will prepare the class path and launch the scala
  repl directly
- `scala.script.name` - a defined property that tells you the name of the top level
  script that is executing - use this to load other resource files that are relative
  to the script itself

See the `test` folder of this repository for a complete example.

## Usage

something.scala

```scala
// sample included script
//

object Something {
  override def toString = "Included script"
}
```

main.scala

```scala
//#include something.scala

println(Something)
```

shell example

```scala
#!/usr/bin/env scala-script

println("scala script")
```

## Outputs

The script pre-processes each scala file and generates a matching `filename.g.scala` file with
the `include` content expanded. When the script is run it is these `.g.` files that are executed,
ultimately, the entire content in the main script. If any source file changes the script
regenerates the `.g.` files as necessary.

The top level script is also persisted as a `jar` file for subsequent launches.

## Caveats

I created this utility for my own use after not finding something suited to my needs. If you find
it useful but need some tweaks please fork it and make the changes you need.

The reason I switch from bash to go was to eliminate the slow startup after the
bash version got progressively slower processing the files.

## License

MIT - See the LICENSE file included in the repository.

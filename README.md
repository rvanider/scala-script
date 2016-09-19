# scala-script

Scala-script is a `bash` shell script that can be used to launch
[Scala](http://www.scala-lang.org) scripts with support for automatic
class path management and a simplified script include mechanism.

## Dependencies

This is a `bash` script so an environment must have bash. There are
no other direct dependencies and the only thing you install is
a short shell script.

- Linux and OSX have been tested
- Scala 2.11.8 has been tested

## Installation

The only file in this repository you need is the `scala-script` shell
script. You can download it and manually install it where and how you
see fit.

    curl -s --fail -L --output "/path/to/store/scala-script" https://github.com/rvanider/scala-script/blob/master/scala-script?raw=true
    chmod +x "/path/to/store/scala-script"

An `install.sh` script is provided that will perform the above and place
the resulting script in `/usr/local/bin`.

    curl -s --fail -L https://github.com/rvanider/scala-script/blob/master/install.sh?raw=true | sh

## Features

- `classpath` - the classpath will automatically be built for any jar files located in the `lib/` folder
  relative to the script being executed
- `include` - other scala scripts can be included using the syntax `//#include path/to/file.scala`

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

## License

MIT - See the LICENSE file included in the repository.

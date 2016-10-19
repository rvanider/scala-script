# Change Log

## 2.1.0

- added support for pass through arguments to the underlying scala
  executable - anything before the script name is passed directly
  to the scala command line
- added a `--nop` flag to turn off the default flags passed to scala
- changed the `scala.script.name` property to use the original name
  of the script instead of the generated file

## 2.0.1

- added the script root folder to the class path

## 2.0.0

- converted to compiled script launcher to improve launch speed

## 1.0.0

- Initial release

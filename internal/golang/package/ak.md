# name: package
# description
package is a code representation of a go package, that can build a source map cache of a package
# features:
 - resolves packages in a module and builds a go structure of the package
 - can render a package source map (compressed what a package represents)
### format:
```
package: <package name>
description: <package description>
features:
  - <feature 1>
  - <feature 2>
file: cmd/akha/backend.go
imports:
types: // used types and interface + assertions too (var _ <interface> = <type>)
exports: // all exports of the package (functions, types, variables, constants)
private: // all private members of the package (functions, types, variables, constants)
```
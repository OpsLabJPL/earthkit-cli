earthkit-cli
============

Command line tools for EarthKit


###Conventions

####Packages

Packages should have at least two files, `types.go` and `package.go`.  All types declared in a package should go in `types.go`.  All package-level functions with no receivers, general utility functions, and variables, whether exported or not, should go in `package.go`.

Functions that have receivers should go into a file named after the receiver type (all lowercase).  For example, if you have a type `foo` then all methods with `foo` as the receiver should be put in `foo.go`.  It is OK to have functions without receivers in the receiver-type source files as long as they are fundamentally related to one of the receiver functions in the source file.  For example, you might have a function that is recursive, but the exported signature is not recursive (to prevent the user from being concerned with the initial case).  It would be OK to include the helper (recursive) function alongside the exported receiver function.  Example:

    func (foo *Foo) RecursiveFunc() {
        doRecursiveFunc(foo, "start")
        return
    }
    func doRecursiveFunc(foo *Foo, initialState string) {
        // ...
        return
    }


###Installation
* Install Go for your platform
* set GOPATH, GOROOT, and PATH according to Go instructions for your platform
* in $GOPATH/src, run: 
    * go get github.com/OpsLabJPL/earthkit-cli
    * cd github.com/OpsLabJPL/earthkit-cli
    * go build
* set up your $HOME/.earthkitrc file replacing your AWS key and secret with those for your own AWS account
* you're ready to run! Run "earthkit-cli" to see the list of available commands and options.
    

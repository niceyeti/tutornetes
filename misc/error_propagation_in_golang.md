# Effective Error Propagation In Go

### TLDR
* Design: by making errors first-class and equivalent to normal return values, Golang avoids the overhead of exceptions (code introspection) and saving state to enter/re-enter from catch blocks. This also make errors first-class parts of api-design, not an afterthought (as with untyped try-catch blocks).
* Errors are defined in terms of their responsibilities to their consumers: stack, location error occurred, and any contextual info.
* Treat errors as part of the exported api of all code
* Use simple, terse error messages to avoid needing to update them when refactored; minimize potential change.
* Preserve error info using errors.Wrap or fmt.Errorf.
* In catch-style code, use errors.Is to evaluate error type.
* Classes of errors may be represented by implementing the Error() interface.
* Integrate errors with log type and level; check log-level at startup to enable/disable certain levels.
* Use fmt.Errof over errors.New
* Use errors.Is instead of ==.

Error propagation in go is an important concept, easily forgotten because of there are multiple ways of doing the same things, per different versions.
In recreational coding it is the sort of thing that is easy to omit in order to prove some other code/system concept, with an "I'll learn it later" attitude. Then when it comes time to write robust production code, you scratch your head and scrape google for as much advice as possible in minimal time.

That is, if you're me.

Errors in a complex system should form a hierarchical tree of possible errors, by wrapping errors in deliberate ways. The difficult part is that errors and their types form implicitly form part of the public interface of a package, but often aren't fully built or poorly implemented. Many libs simply return raw errors, without wrapping or hierarchical semantics.

Robust error handling implementations have the following requirements:
* XOR: best practice should let the caller assume that a function returning a non-nil error should ignore other returned values
* messages should be lowercase messages
* messages should be recursively composable:
    * `errors.New("file not found")`
    * plus `errors.New("file system error")`
    * equals output: `"file system error: file not found"`
* Wrap errors using `fmt.Errorf("FuncName %w", err)`
    * `failed finding or updating user: FindAndSetUserAge: SetUserAge: failed executing db update: connection error`
    * IMPORTANT: wrapping like this is compatible with `errors.Is` and `errors.As`
    * As a software pattern, wrapping is primarily good for adding contextual traceability, like the func name in which an error was detected, or other vars.
* On the other hand, it is best not to clog up your public interface with error definitions, since users will then depend on those definitions and they cannot be changed.

As part of the public interface of a package:
* publicly declared errors allow well-defined semantics, using type switches
    * `type someErr struct {} .... implements Error()`
* however, simple errors are much simpler than implementing the Error interface:
    ```
    package some_package
    var ErrSomeError = errors.New("aw crud")
    // elsewhere:
    if err == errors.Is(ErrSomeError) {}
    ```
* Use a tagless switch to handle these defined errors, effectively reducing/closing the number of unknown error paths:
```
switch {
    case errors.Is(ErrFirstClassError):
        // handle it
    case errors.Is(ErrOtherError):
        // handle a different error
    default:
        // handle unknown errors
}
```
    * Note that implementing the Error interface would then allow more structured error handling, by passing data along with the error.
* Using `errors.Is()` is preferred over `==`

Useful interfaces:
* `errors.Is`: Unwraps an err successively and returns true if the error is the INSTANCE of a passed error.
    ```
        if errors.Is(err, fs.ErrNotExist) {
    ```
* `errors.As`: the same as `Is` but unwraps and returns the passed error if it satisfies the underlying type of passed error. Note the distinction, since this effectively detects more structure errors:
    ``` 
        var pathError *fs.PathError
		if errors.As(err, &pathError) {
    ```
* Use errors.Unwrap() to unwrap errors, and fmt.Errorf("%w", err) to wrap them.

























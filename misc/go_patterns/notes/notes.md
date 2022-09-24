# Golang Design Pattern and Architecture Notes

*We build our computers the way we build our cities—over time, without a plan, on top
of ruins*

\- Ellen Ullman, The Dumbing-down of Programming

## Lesson 1: Open Closed Principle
Specification pattern:
Although I would just pass in a function for filtering, the instructor's example is that of filtering
a collection by:
1) Implement a higher level component called "Filter"
2) Filter defines a single method FilterBySpec(s *Spec) that takes a specification describing a specific filtering operation
3) Define Spec objects to filter by different properties of some base data objects: filter by color, size, etc.

This is the same as the interface for Sort in the standard library: pass in interface defining the filtering predicates.

The example demonstrates the open-closed principle since the interface of Filter is open to extension (by adding more predicate implementers)
but the filter behavior/code will never need to be changed. Filter predicate implementers can be managed separately.

By contrast, passing in a `func(*Foo) bool` method is the most robust and extensible method, probably preferred, but I can see the
author's point about extension.

## Lesson 2: Liskov Substitution Principle
Any subclass should be substitutable for the superclass from which it is derived.

The instructor demonstrates how implementing a Square atop a Rectangle interface is problematic
because it breaks the previous invariants of the code, thus violating the Liskov Substitution Principle.
For example, a Square with a SetHeight() method that sets both height and width would invalidate
any previous values stored from GetWidth(), thus breaking the invariant that values read from 
this object do not change unexpectedly:
```
width := square.GetWidth()
square.SetHeight(123) // also sets width!
// Now this test fails!
So(width, ShouldEqual, square.GetWidth())
```

## Lesson 3: Interface Segregation Principle
Basically, try to break up interfaces into the minimal set of interfaces that are least likely to change.
The instructor's example is that of an OldPrinter interface containing Scan, Print, and Fax methods.
These should instead be broken up into Scanner, Printer, and Faxer interfaces, and gathered into one
for instance, by embedding.

Basically just prefer high granularity, but without proliferating too many interfaces or over-abstracting.

In Golang, there are a few rules of thumb:
1) Return structs, consume interfaces
2) Be conservative about writing interfaces; if there is only a single implementer of an interface, and only one user (e.g. your own app) then there is little reason for an interface.

## Lesson 4: Dependency Inversion Principle
High level modules should not depend on low-level modules.
This is consistent with Uncle Bob's onion-style architecture diagram in Clean Architecture: high level modules should not depend on low level modules.
An example violation of this principle was golang's x509 library implementation of SystemCertPool() requiring network calls/communication on the windows platform:
such a library should not depend on networking and external resources, but this was required by windows behavior (and ultimately not merged for that reason). Put simply, high level modules should not require changes based on their access to lower level modules and their idiosyncracies/volatility.

A more direct example is a high-level module like Research containing (depending on) a field for low-level
collection objects directly. In the simplest sense, the dependency could be converted to a getter implemented
on the low-level module.

Low-level: data, db, connections and file storage, etc.
High-level: behavior, application logic.

The motivation of this, for instance, is that the high-level module does not need to change when any low
level module is uses is modified or extended.

*YAGNI*: you ain't gonna need it. Minimize interfaces so implementers need to do the minimal amount of work needed.

## Lesson 5: Builder Pattern
Piecewise object construction using a succinct api.
* strings.Builder{} is a very simple example, allowing one to construct a string through a series of calls.
* Behavior: api for building up something incrementally

Uses:
1) Build things incrementally
2) Build something whose composition has multiple domains and domain-specific requirements, using builders for each domain
3) Develop a fluent interface to build up validation of an object.

Problem: you have an object that you want to build up in steps, like strings.Builder.
You don't want users of the api to be aware of the underlying structure, such as an html ele tree.
Implement the internals of html (indentation, closing tags, etc) inside this builder, HtmlBuilder.
* The builder api defines the semantics of the user's task: add thing, add another thing, etc
* Underneath, the builder defines the underlying semantics
Thereby these two structural definitions are separated.

Builder often uses fluent interfaces, aka method chaining.

When to use Builder:
1) You want to build something complicated in an incremental way, using method chaining
2) There are two behaviors to separate: the logical/application logic and the underlying structure that you want to hide from the user of the builder.
3) There is a domain-specific language (DSL) or multiple DSL's defining how an object should be built (e.g. html, xml, or json string builders).

(3) is where things get fancy, since you can facet off of builders, say a PersonBuilder to a PersonAddressBuilder and PersonJobBuilder.
This allows using multiple builders for separate domains to build a single entity.
These patterns are especially useful/clean for building validation logic: authentication, authorization, etc.

To clarify via an example, the high level code looks like this:
```
pb := NewPersonBuilder()
pb.Lives()  // Top level call to pivot to the address-builder
    .At("123 Fake St")
    .In("Happyville")
    .WithPostcode("8675309")
   .Works() // Top level call to switch to the job-builder
     .At("MegaCorp Inc")
     .As("Mop mechanic")
  person := pb.Build()
```
Pivoting between the builders is done using struct embedding of the base builder into the domain-specific builders.

Modifications to the pattern:
1) You can instead pass in actions whose parameter is the object, and build a set of actions incrementally. Then when the
object is needed, apply all of the actions. This allows not having to implement new builders for new actions, as well as
lazy construction.

Code strategies in go:
1) Embed builders in other builders to allow pivoting off their behavior
2) Build fluent interfaces by returning the builder from each method

## Lesson 6: Factory Pattern

In golang, most of this is built-in or just idiomatic: declare interfaces and create objects using NewFoo() that return a private struct implementing
a public interface. The only extension described was returning a function by which to construct something later or elsewhere:

```
type Employee struct {
    Role string
    Salary int
    Name string
}

// pseduo code
func NewEmployeeFactory(role string, salary int) {
    return func(name string) {
        return &Employee{
            Role: role,
            Salary: salary,
            Name: name,
        }
    }
}
```

The above pattern allows passing around factories, parameterizing them more flexibly according to when they are used, et cetera.
But also seems pretty esoteric, unnecessary except for special cases. Used in middleware and endpoints construction; but note 
I'm referring to the construction-time pattern, not the recursive pattern of returning endpoint functions, since that's closer
to Chain of Responsibility.

## Prototype Pattern

This mostly covers copying:
1) implementing a DeepCopy method manually
2) implementing DeepCopy using binary serialization with encoding/glob

The Prototype pattern just implements a behavior whereby you implement a "New" or other factory pattern that takes some
pre-set object and its values, then modifying them as needed.

A key point is that you can share the types between different implementations and expose only different methods for 
the different behavior. For instance an Employee type underneath, with some prototype base definition, exposed
through a NewMainOfficeEmployee and NewRemoteEmployee.

## Singleton Pattern
Constrain something to a single instance.

Implementation:
1) Define a single pointer var in global package space
2) Use sync.Once to initialize it

Code smells:
Singleton Pattern violates the dependency inversion principle by making the high-level module depend on the low-level details of a lower level
module, such as a database or anything else that you would initialize in sync.Once. Often very low-testability.

So to implement Singleton successfully, make sure you are not violating the dependency-inversion principle (DIP). This will be most
clear by writing unit tests and letting their requirements drive the refactoring required to minimize requirements on the low-level details 
of the underlying singleton data, behavior, etc.
1) Use sync.Once and var to initialize singleton; package-level `init()` may also be used, since it has ordering and threading guarantees.
2) Use DIP to refactor Singleton using test-driven development

## Adapter

Use Adapter to join two interfaces/apis: 1) the interface you're given 2) the interface you have (e.g., your legacy codebase).

An Adapter is really just any function like `VectorToRasterImage(img *VectorImage) *RasterImage`, except exploded out using
interfaces and private structs to maintain multiple definitions or other adapters, etc. But its really quite simple, an Adapter
is just a conversion function, hidden through the use of (over-engineered) interfaces to define those conversion responsibilities separately.

How often is it needed? Probably not much; conversion is often one-off code because of the O(n^2) potential relationships between
objects, and little need for generalization beyond one or two cases of conversion. But you might want it for encapsulation, if 
an adaptation is quite complex or has some additional maintenance requirements.
* Maintenance requirements
* Data (object to object) complexity
* Protocol conversion
* To encapsulate a single set of responsibilities, and minimize resources consumption, for object conversion

## Bridge

Solves a cartesian product complexity problem. Say you have a set of shapes, Square and Circle, and a set of renderers, RasterRenderer and VectorRenderer.
Instead of writing a SquareRasterRenderer, SquareVectorRenderer, etc, as a 2x2 set of implementations, you refactor to ensure a minimal number of implementations.

This is less complicated than described: the solution is to add a Rendered interface as a member of Circle and Square, then call the appropriate renderer.
All this amounts to is looking at the structure of your dependencies and minimizing them by decoupling and composition. It really is simple, and is a rote task
of examining one's object hierarchy and looking at cleaner ways of decoupling using encapsulation.

## Composite 

A Composite is any object with a recursive relationship with members of its own type:
```
type Foo struct {
    Data []byte
    Children []*Foo
}
```
This can be used for layering things, like neurons (or layer) in a neural net, records in a database, or semantic language features like a parse tree.
But you can see how such a data object is amenable to specifying behaviors that can be performed on such a representation: DepthFirst(), BreadthFirst(),
iter(), and so forth, for iterating or searching through such objects.

* Composites can also be bidirectional, such as defining both Parents and Children fields. This further generalizes to other multi-dimensional representations.

## Decorator

Aggregate the behavior of multiple underlying implementations behind a simplified interface.

Example: say you have a Bird and Lizard struct, each with Fly() and Crawl() methods respectively, and want to implement a Dragon.
1) Implement a Dragon struct embedding Bird and Lizard, since it both Flies and Crawls.
2) Use interfaces to define the common behavior and conceal the unnecessary parts.

Decorators rely on composition and concealment.

## Facade

Provides a simple interface over a large, complicated body of code.
The Facade provides a simple, self-descriptive representation of a complex object.
It is often just a loose wrapper defining a set of coherent behaviors.
A such a Facade may still allow access to the underlying lower-level type definitions to allow direct modification of those members.
The course gives an example of a ViewPort facade over a Buffer implementation, where a ViewPort is just a sub-region of a Buffer.
For instance a ViewPort may have a state definition, such as the current offset, which is applied to the Buffer.

```
func NewViewPort(buf *Buffer) (*ViewPort) {
    return ViewPort{ buf: buf }
}

func (vp *ViewPort) GetCharAt(index int) {
    return vp.buf[vp.offset + index]
}
```

In this manner, a Facade provides a wrapper around some more complicated underlying implementation, and may encode some state that simplifies dealing with it.
It need not conceal the underlying/encapsulated types, but may still allow the user to use these directly, using the Facade when it suits them.

## Flyweight

A space optimization technique separating storage of data from the objects using it. String interning is the simplest example:
* A string in golang is a pointer to data and length
* Since strings are immutable, multiple strings can have a pointer and length comprising the overall storage underneath

This is really a matter of data refactoring: for example instead of storing a bool array 'capitalize' matching the 
length of some string s, and describing letters to be capitalized, simply stored the start and end position of capital letters.
I believe there was a good example of this in my old C++ n-gram code; also tries (read-maps?) from Ananth's course.
Inverted indices in search engines are another example.
For example, mapping words to indices, converting input streams to mere integers, and using these integers for O(1) array lookups
rather than using some complex underlying data structure. This is a classical example:
1) implement a []string array
2) implement a []int array index-aligned with (1)

Manipulate each of these to translate a word to its integer index, and back again, in turn. This pattern occurs often for sequential or graphically-defined data using multimaps, for which you don't want to store multiple copies of unique objects, such as words; instead you encode these as numbers in some other lookup, then store only the numbers. In many cases the scale of data can then be encoded in mere arrays instead of complex trees.

Flyweight usually involves a few tricks:
1) an inversion principle for describing a property of data (integer, index, pointer, etc), rather than storing it everywhere
2) temporary objects when iterating or accessing the description
3) The idea of ranges of some property over a homogeneous collection

Tries may be an example of a flyweight, but here you see the distinction between software engineering, code, and algorithms; the rules
of each need not be the limit of their description, how they are used is what matters.

## Proxy

A type that functions as an interface to a particular resource.
Used when you pass in some dependency to a factory, e.g. NewFoo(bar *Bar), and extend the behavior of the dependency with additional behavior: access control, etc.

Virtual proxy: pretend you have a proxy when you don't, which is useful for laxy loading/initialization.
A proxy usually has a proxied object, e.g. a pointer to the thing being proxied.
It extends it by, say, not creating the thing until someone actually tries to access it.
A Virtual Proxy is "virtual" simply because the thing hasn't been created yet.

```
type LazyFoo struct {
    filename string
    bar *Bar // the thing being proxied
}

func (lf *LazyFoo) Do() {
    // lazily instantiate a Bar
    if lf.bar == nil {
        lf.bar = NewBar(filename)
    }
    // do stuff...
}

func NewLazyFoo(filename string) {
    return &LazyFoo{filename: filename}
}
```

Comparison with Decorator:
* Proxy tries to mimic the identical interface of the thing being proxied
* Decorator may represent a heterogeneous set of underlying things behind a single 
* Proxy is more about extending (lazy init, authorization, etc), whereas Decorator is often for aggregation


## Chain of Responsibility

A chain of components responsible for processing some item/change, any of which may terminate processing at a given point.
* recursively defined and built
* useful anywhere a chain of logic is needed (e.g. middleware)
* useful for in-memory representations of things, a pure functional description of items (see Creature example below)

An example interface exposes an Add(*Foo) method to add another modifier, and a Handle() method to apply it.
Concrete implementations maintain a linked list of these.
* Modifiers are recursively defined, forming a chain through their 'next Modifier` member.

```
type Modifier interface {
    Add(Modifier)
    Handle()
}
```

The Modifier interface forms a functional graph of elements for processing changes.
Likewise, any function can avert further processing by simply not calling next.Handle().

Chain of Responsibility lends itself to more complex enterprise patterns whereby 'Modifiers' can be defined 
and instantiated to handle events and return particular views of an object in memory. In the course example,
Modifiers may be created to modify a Creature's health values, such that the modified vaues are not stored,
the modifiers just return this view of them. Whether or not the values are persisted is not super important,
simply the decoupling that allows modifiers to be instantiated, defined, and managed independently, using a common interface.

## Command Pattern

Gist: Command is useful anywhere that actions are first-class objects to be saved/recorded.
To implement, just think of the requirements for implementing a symmetric Undo() action and work backward.

Commands provide:
* serialization, recording/saving, and undo-ability of actions
* macros, delayed execution, or other forms of control
* useful when persisting or maintaining a record of transactions or otherwise abstracting commands as first-class objects

Patterns:
1) Another processor component executes the command
2) The command executes itself

```
type Command interface {
    Call()
    Undo() // make each Command symmetric
}

type Action int
const (
    Deposit Action = iota
    Withdraw
)

type BankCommand struct {
    ba *BankAccount
    action Action
    amount int
}

func (bc *BankCommand) Call() {
    switch(bc.action) {
        case Deposit:
            bc.account.Deposit(bc.amount)
        case Withdraw:
            bc.account.Withdraw(bc.amount)
    }
}
```

The example above is very simple. Modifications to the Command pattern allow the developer to determine
what should be a part of the Command, how it should reference the thing that processes the command (if it does
reference it), and so on. Other modifications accomodate extended behavior, such as undoing and recording:
* add a `succeeded bool` field to each command to allow undoing the command along with symmetric commands to undo each action defined
* add serialization, etc.

Higher-order modifications can be implemented as well using the Composite Command pattern (aka Macros), consisting of multiple
commands. Composite transactions occur naturally, such as needing an action to transfer money between accounts: withdrawing from one and
depositing in another.

```
type Command interface {
    Call()
    Undo()
    Succeeded() bool
    SetSucceeded(value bool)
}

type CompositeCommand struct {
    commands []Command
}

// implement all of the Command interface 
func (cc *CompositeCommand) Call() { ... 
func (cc *CompositeCommand) Undo() { ... 
// ... etc
```

Composite commands need to simply satisfy the Command interface and obey their same invariants/requirements: symmetry, all or nothing behavior, etc.
For example if one command in a sequence fails, the previous commands must be undone. This is a good way to encode any hazards, such as the possibility
of cascading failures, ensuring undoability (even when Undo itself fails!) using checkpoints, and things like that.

Examples: see the Cobra project for a terrific example of building CLI applications atop the Command pattern.

## Interpreter Pattern

The Interpreter pattern occurs anywhere you have an interpreter: compilers, html/xml parsing, etc.
It consists of two steps:
1) lexing: converting input into tokens, elements, or other atomic descriptions of the text input
2) parsing: developing a hierarchical representation of the elements from (1), an abstract syntax tree

The course gives an example implementation of a calculation interpreter:
1) lex input like "(2 + 13) * (4 + 5 * 9)" into a sequence of atomic components
2) parse these atomic components into BinaryOperations and UnaryOperations

These can then be processed elegantly. Error handling can also be implemented in a manner to push up notifications
from each stage, such as a language server notifying a user of a syntax error in code.
Parsed data can then be processed using the Visitor pattern.

## Iterator Pattern

Generators can be written idiomatically using go-routines and channels using the common producer/generator pattern:
```
func (f *Foo) BarGenerator() <-chan string {
    ch := make(chan string)
    go func(){
        defer close(ch)
        for _, s := range someStringSlice {
            ch <- s
        }
    }
    return ch
}
```

Traditional iterators store state (where we are in a collection) and a pointer to the current item, and expose a MoveNext() method.
An Iterator can implement a specific kind of iteration, such as in-order/post-order/pre-order in trees, implemented as:

```
type Iterator interface {
    MoveNext() *Node
}

type Node {
    left, right parent *Node
    value interface{}
}

type InOrderIterator {
    current *Node
    root *Node
}

func NewInOrderITerator(root *Node) Iterator {
    it := &InOrderIterator{root, root}
    it.Reset()
    return it
}

func (it *InOrderIterator) Reset() {
    it.current = it.root
    for ; it.current.Left != nil; {
        it.current = it.current.left
    }
}

// Note: assume we have initialized iterator to leftmost node in the tree.
func (it InOrderIterator) MoveNext() *Node {
    if it.current.Parent != nil {
        it.current = it.current.Parent
    } else if it.current.Right != nil {
        it.current = it.current.Right
    } else {
        it.current = nil
    }

    return it.current
}

```

Iterator is not natively amenable to golang, primarily since the `range` keyword cannot be applied arbitrarily
to items implementing some iterable interface. But obviously a simple loop works just fine.

## Mediator Pattern

A mediator facilitates communication among components without them necessarily having to be aware of one another.
NATs seems kind of mediator'ish, though strictly it is PubSub; Reactive extensions are another example.
Used for volatile objects/participants in some activity who need not hard links to one another: chat participants, gamers, etc.
The foremost elements of the Mediator pattern are:
1) Participants/objects need not be aware of eachother
2) Relationships are graphically defined
3) Message passing occurs in some form

The Mediator itself may be very sparse, much like NATs, whereas the users of the Mediator fulfill much of the pattern
by implementing their behavior such that they don't need to reference one another.

Any time you have graphical relationships that can be organized for separation, size, etc., you should
think of the Mediator. It is essentially a directed graphical description whereby participants can receive
messages or send them as broadcast or private.

## Memento Pattern

Memento is simply a 'Snapshot', often with no methods at all, storing the complete state of a system in a convenient and restorable manner.
Git is kind of a Memento machine of sorts, and an example of restoring/creating snapshots.
Not required to implement Undo/Redo or other restoration aspects; it is a mere data object.
* A snapshot data object
* Has no methods
* Used to implement Undo/Redo elsewhere in a system

```
type Foo struct {
    state int
}

type Memento struct {
    state int
}

func (f *Foo) Bar(newState int) *Memento {
    f.state = newState
    return &Memento{newState: f.state}
}

```

Usually accompanied with Undo/Redo methods.

## Observer Pattern

The Observer pattern is an event-based pattern consisting of an observer and an observable,
and usually involves a framework glueing them together.



## State Pattern

An object's behavior changes when its internal state changes.

Pedantic implementation: the state machine has-a state; individual structs implement each state internally, and
encode the transition model, by reference the state machine, creaitng the new target state, and setting it
on the state machine

Realistic implementation: instead of structs and many objects, just implement a state machine as a graphical
model using a map from states to their set of outgoing transitions ("triggers"). 

```
type State int 

const (
    Init State = iota
    SomeState
    AnotherState
    EndState
)

func (s State) String() {
    switch s {
        case Init: return "Init"
        case SomeState: return "SomeState"
        case AnotherState: return "AnotherState"
        case EndState: return "EndState"
    }
}

type Trigger int 

const (
    Foo Trigger = iota
    Bar
    Baz
)

// Implement trigger stuff here...

type TriggerResult struct {
    Trigger   // action
    State     // target state
}


var model map[State][]TriggerResult {
    Init: {
        {Foo, Init},
        {Bar, BarState},
    }
    // the rest of the state/transitions here...
}
```

I would think that the State Pattern should not cross many boundaries or be too unwieldy, since it can resemble
over-engineering and could become unwieldy. Much better to have small state machines defined in a single place.
Even using a bunch of const int states and transitioning between them using a for-switch loop would work well in practice.

## Strategy Pattern

Gist: an abstraction for algorithms, whereby the algorithm can be completely abstracted, and even broken down into components that
can be varied/swapped at runtime. A Strategy can simply be defined as an interface whose methods are the common components
of some algorithm: a step by step process, a string builder (html, markdown, etc), or some other processor.

1) Implement the verbs of the Strategy in an interface; these are the core components of some algorithm
2) Users can then swap out implementers of the interface

Example:

```

type ListStrategy {
    Start(strings.Builder)
    End(strings.Builder)
    AddListItem(strings.Builder, string)
}

type MarkdownListStrategy {}

func (md *MarkdownListStrategy) Start(build strings.Builder) {
    build.Write("<ul>\n")
}

func (md *MarkdownListStrategy) End(build strings.Builder) {
    build.Write("</ul>\n")
}

func (md *MarkdownListStrategy) AddListItem(build strings.Builder, item string) {
    build.Write("* " + item + "\n")
}

```

Some higher level textual object can then contain instances of the ListStrategy which it swaps out according to what is needed: html, markdown, etc.

## Template Method

* In go 1.18+ one can simply write templated methods!


In the course, the author defines a Template method as simply a thing that calls the components of some other interface in turn.
So it 'templates' some behavior by structuring how the calls are made.
For example:

```
type Game interface {
    Start()
    HasWinner() bool
    TakeTurn()
}

func PlayGame(game Game) {
    game.Start()
    for ; !game.HasWinner(); {
        game.TakeTurn()
    }
    fmt.Println("Game over!")
}
```

Above, the Template method `PlayGame` simply calls the behavior the Game interface, 'templating' it in terms of defining its generic behavior.
The example seems oversimplified. A more contrived but elucidating example would be a PlayGame method that took all
of the game methods as dependencies; in this manner it would 'template' some behavior but allow the individual functions to define their own behavior.

## Visitor Pattern

Intrusive pattern: the intrusive pattern requires changing the interface of components in a hierarchy, violating OCP.
But to do so, just pass in the dependency (the 'Visitor') to some function implemented by all components.
This scales poorly, because to implement different 'Visitors' in this manner one adds the methods to the component interfaces.

A 'Reflective' Visitor instead takes in the component and defines the work and traversal pattern.
Dmitri's example simply uses a type-switch. It is still somewhat inflexible because of the hardcoded 
type checking, etc.

```
type Expr interface{}

type DoubleExpr struct {
    value float64
}

type BinaryExpr struct {
    right, left Expr
    operator string
}

func PrintFoo_Visitor(expr *Expr, sb *strings.StringBuilder) {
    // NOTE: this pattern is bad because new type requires revisions to this switch
    switch v := expr.(type) {
        case DoubleExpr:
            sb.Write(v.value)
        case BinaryExpr:
            sb.Write("(")
            PrintFoo_Visitor(v.left)
            sb.Write(v.operator)
            PrintFoo_Visitor(v.right)
            sb.Write(")")
        default:
            panic("Aw crud! new type not implemented!")
    }
}
```

Double dispatch def: separates concerns by having components implement a Foo(*Bar) method, taking in the Bar
dependency and then calling the appropriate method on Bar, to which they pass themselves. Though confusing, there is not
much going on here accept to ensure that the code depending on a Foo object receives the Foo object, but lives and is
maintained wherever Foo is being manipulated and used. A little awkward, but allows things to live and be maintained
according to their responsibilities.
* Note that double-dispatch is still slightly intrusive, since it requires adding the Accept method.

The 'classic' Visitor uses double-dispatch instead, which uses a layer of indirection to delegate objects and make all of the types/interfaces agree:
1) Components implement an Accept(visitor *Visitor) method as part of their interface.
2) In the Accept method, the component passes itself to the hard-defined method on the visitor's interface.
See the code example; pay attention to simply how double dispatch works, whereby a top-level Visitor interface calls
specific visit implementations, and the visited objects simply implement an Accept() method to pass their info across.
It is somewhat over-engineered and doesn't quite fit the open-closed principle since modifications require updating all implementers.
However it does ease the implementation of implementing new visitors; so this is a robust pattern when you need to visit 
some hierarchical information in some manner with different logic/behavior (e.g. printing an expression vs evaluating it).
So it seems like this works well in contexts where the structure of information is unlikely to change much.

The key is:
* a top-level inerface defines all the methods for visiting
* Visitors implement these interfaces on different components in the hierarchy
* components themselves implement an Accept(ev TopLevelVisitor) method to pass themselves into its methods, e.g. ev.VisitFoo(self)

## Futures

A future allows passing around some item of work until its result is needed by calling .Result().

Some of the details below are inaccurate/incomplete. The gist is simply that the InnerFuture receives a channel on which the output of a long operation will be sent. The Result() implementation uses a sync.Once function to ensure it is only called once, which forces a wait on the result of the work channel and/or errors.

I tend to forget the necessity of the sync.Once requirement, to prevent Result() from being called more than once.

```
type Future interface {
    Result() (string, error)
}

type InnerFuture struct {
    input <-chan input
    errs <-chan error
    res chan string
    once sync.Once
}

// Result blocks until the result is returned.
func (inner *InnerFuture) Result() (result string, err error) {
    inner.once.Do(func(){
        wg := sync.WaitGroup{}
        
        wg.Add(1)
        go func() {
            defer wg.Done()
            result := <-inner.input
            err := <-inner.errs
        }()
        wg.Wait()
    }

    return
}

func SlowFunction(ctx context.Context) Future {
    result := make(chan string)
    errs := make(chan errors)

    // purely demonstrative of a slow operation of some kind
    slowOp := time.After(5 * time.Second)

    go func() {
        defer close(result)
        defer close(errs)

        select {
            case res, ok := <-slowOp:
                if !ok {
                    errs <- ErrInputClosure
                    return
                }
                result <- res
            case ctx.Done():
                errs <- ctx.Err()
        }
    }()

    return InnerFuture{input: in, errs: errs}
}

```


## Cloud Native Patterns

The chainable CircuitBreaker pattern allows returning a function that defines some state, such
as retries, throttling, circuit breaking, and so on. This provides much more compact form of stored state than defining a receiver.
The top level nested calls are the important part:
```
    func myFunction func(ctx context.Context) (string, error) { /*...*/ }
    wrapped := Breaker(Debounce(myFunction))
    response, err := wrapped(ctx)
```

Full CircuitBreaker example follows. Note how returning a func allows defining some state
for it, and understand how it must be called. The best pattern wouldn't put this in app or library
code, but implemented sidecar proxies and driven by config (e.g. a service mesh like Istio). The requirements are:
* context all the way down
* concurrency safe (func doesn't know who will call, or how many times)
* chainable

```
type Circuit func(context.Context) (string, error)

func Breaker(circuit Circuit, failureThreshold uint) Circuit {
	var consecutiveFailures int = 0
	var lastAttempt = time.Now()
	var m sync.RWMutex
	return func(ctx context.Context) (string, error) {
		m.RLock()
		// Establish a "read lock"
		d := consecutiveFailures - int(failureThreshold)
		if d >= 0 {
			shouldRetryAt := lastAttempt.Add(time.Second * 2 << d)
			if !time.Now().After(shouldRetryAt) {
				m.RUnlock()
				return "", errors.New("service unreachable")
			}
		}
		m.RUnlock()                   // Release read lock
		response, err := circuit(ctx) // Issue request proper
		m.Lock()
		defer m.Unlock()         // Lock around shared resources
		lastAttempt = time.Now() // Record time of attempt
		if err != nil {
			consecutiveFailures++
			return response, err
		} // Circuit returned an error,
		// so we count the failure
		// and return
		consecutiveFailures = 0 // Reset failures counter
		return response, nil
	}
}
```

## Other Patterns and Architecture

1) Return structs, consume interfaces: https://www.reddit.com/r/golang/comments/t9no58/clean_architecture_best_practices_in_go/
    * This is good for general decoupling, but in particular for dependency injection and testing. This is done by having ones public api consume an interface defined as a private interface in one's own package; as long as the passed interface implements the required methods, it is valid. But most importantly this eases testing, and mocking frameworks like testify and gomock are built directly around it.
2) 3-Layered Architecture: db, business, and API.
    * Each layer can be maintained and modified independently of the others
    * Presentation: css and js for example; the frontend client app
    * API: the api with which the client app interacts
    * DB: the backend database for the application

## Go Tricks and Lessons Learned

1) Struct embedding can be used for mock inheritance, whereby a derived thing
optionally calls its base behavior or overrides it:
    ```
    type FooBar interface {
        Do()
    }

    type Bar struct {}

    func (b *Bar) Do() {
        fmt.Println("Bar Do called")
    }

    // Also implements Do() interface 
    type Foo struct {
        Bar
    }

    func (f *Foo) Do() {
        fmt.Println("In Foo Do()")
        // Call the base behavior as well
        f.Bar.Do()
    }
    ```

2) Use faceting to pivot between domain-specific languages for an object, such as building
aspects in different domains: a person's address, personal attributes, car, etc. See
the builder-facets code example: the facets `Works() *PersonWorkBuilder` and
`Lives() *PersonAddressBuilder` each construct and return builders for a domain. 
However since each embeds a base PersonBuilder, this allows pivoting between
the top-level Works/Lives builders fluently.

3) Use the builtin lib List library. It also has priority queues and such.


## gRPC

gRPC is merely a binary RPC format, the code for which is generated using an external tool.
The process:
1) Define .proto files for one's DTO type definitions, each field marked by field numbers. Yes, these are tightly coupled to the corresponding Go definitions, but there is probably some organizational mechanism to link them somehow. Perhaps a compile-time version-string comparison somewhere, like interface satisfaction checks: `var _ RequiredInterface = (MyType*)(nil)`.
```
syntax = "proto3";

option go_package = "github.com/cloud-native-go/ch08/keyvalue";

// GetRequest represents a request to the key-value store for the
// value associated with a particular key
message GetRequest {
    string key = 1;
}
...
```
2) Define the service with 'rpc' prefix. This could be done in the same .proto file:
```
    service KeyValue {
        rpc Get(GetRequest) returns (GetResponse);
        rpc Put(PutRequest) returns (PutResponse);
        rpc Delete(DeleteRequest) returns (DeleteResponse);
    }
```
3) Compile using protoc. Specify the source and dest folder (often both are the app/src directory) and output language `--go_out`.
```
protoc --proto_path=$SOURCE_DIR \
--go_out=$DEST_DIR --go_opt=paths=source_relative \
--go-grpc_out=$DEST_DIR --go-grpc_opt=paths=source_relative \
$SOURCE_DIR/keyvalue.proto
```

The rest is glue-code. See Cloud Native Go for an example.
The benefits are:
* binary speed
* procedural RPC interfaces


## Templates

I more or less know templates after using them to create views.
See view code in github.com/niceyeti/tabular server.

## Plugins

The requirements for Golang plugins are minimal: 
* define a main package in the plugin
* define one or more exported variables or functions, satisfying some interface contract:
```
    var Animal duck
    // Assume Says is part of an interface defined in the plugin-user code
    func Says() 
```

The calling code uses the builtin plugin package to load plugins, search for symbols,
and call functions satisfying specific interfaces.

There are additional glue-code specifications: plugins are built as object files, callers
go through a process of searching plugins for symbols/functions using the plugin package, etc.
The benefit is that plugins can be added simply without recompiling the entire project,
extending functionality by simply adding and building a new plugin. For instance, in ABLE
one could define page parsers as plugins, specify via url/source, and load them on the fly.
* Upside: separate compilation
* Downsides:
    * it may be appropriate to decouple by implementing functions via gRPC in another process,
such that panics in plugins don't crash the requesting process.
    * version specific contracts (this actually seems like a huge downside)

Decoupling: see Hashicorp's gRPC go plugin library, which operates go plugins in separate processes
via gRPC. Even if not used, its useful food for thought. Where this might be useful is with plugin sidecars running within the same pod context (security, resources, other specs). It has a nice team-separation property, in that plugins can be maintained and controlled (devops) independently.

## Hexagonal Architecture

Hexagonal architecture is a cloud realization of Uncle Bob's onion architecture, with its dependency rules and boundaries, such that all dependencies point inward: innermost are stable business objects, followed by business logic, followed by data adapters and connections (db, net, etc), followed by the actual drivers. Hexagonal architectures use the same layout, more loosely defined:
* domain objects
* business logic
* ports/adapters: these know how to transform data from external actors and ferry it into the core logic

Example app organization:
```
├── main.go       // glues dependencies together 
├── core
|    └── core.go   // core logic, contains only stdlib dependencies
├── frontend       // api components (adapters) which depend on core pkg
│   ├── grpc.go    // grpc definitions
│   └── rest.go    // rest api
├── transact
│   ├── pglogger.go    // postgres logger definition
│   └── filelogger.go  // file logger definition
```

## Cobra

Use flags wherever possible for simple programs, but cobra provides a nice interface for building more complex CLI applications
centered around Command patterns. Use when tasks can be organized as commands, including subcommands,
such as for command line automation.

## Viper

Viper is a companion to Cobra, the command line utility, but deals with configuration.
It brings in many dependencies, hence is only suited to large systems projects.

AWESOME: Viper has builtin functionality for reading, watching, and even remote-fetching of config files.
It even natively support watching etcd. Thus can watch for config file updates, fetch them from a remote store,
and surely support kubernetes as well (though k8s can provide dynamic config-map updates via config-map volumes).















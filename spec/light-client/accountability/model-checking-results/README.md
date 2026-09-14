# Results of checking inductiveness and safety

This file summarizes the computational results of model checking inductiveness of `TypedInv`
and safety of `Agreement` and `Accountability`. This is done for `N=4`, `T=1`, `F=1`.

The checks are performed with [Apalache][] v0.58.3.

## The inductive invariant preserves agreement

```sh
apalache-mc check --length=0 --cinit=ConstInit --init=TypedInv --inv=Agreement MC_n4_f1.tla
...
The outcome is: NoError
Total time: 286.655 sec
```

## The inductive invariant preserves accountability

```sh
apalache-mc check --length=0 --cinit=ConstInit --init=TypedInv --inv=Accountability MC_n4_f1.tla
...
The outcome is: NoError
Total time: 299.306 sec
```

## The inductive invariant is satisfied in the initial states

```sh
apalache-mc check --length=0 --cinit=ConstInit --init=Init --inv=TypedInv MC_n4_f1.tla
...
The outcome is: NoError
Total time: 4.84 sec
```

## The inductive invariant is preserved by transitions


```sh
apalache-mc check --length=1 --cinit=ConstInit --init=TypedInv --inv=TypedInv MC_n4_f1.tla
...
The outcome is: NoError
Total time: 11329.309 sec
```

[Apalache]: https://apalache-mc.org/

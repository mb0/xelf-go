Generics
========

Most of the code was initial written when generics where not yet on the horizon.
Although proxy literals cover all the cases that the "generic" literals don't, we might want
to revisit the container types list and dict for additional type safety and the number literals to
maybe drop number proxy literal all together.

A generic number type `type Num[T constraints.Integer] T` is
[not allowed in go 1.18](https://github.com/golang/go/issues/45639)

But we could change the int and real proxies to use type parameters, maybe to avoid reflect calls?
Same goes for list and dict proxies, we can potentially avoid calling into reflection.

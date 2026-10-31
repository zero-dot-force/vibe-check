"""Classification and cohesion scenarios for the ripwire adapter.

Each declaration in this module isolates exactly one scenario the
adapter must classify or measure (see FIXTURES.md at the testdata
root for the full fixture manifest):

- Cohesive: three methods sharing self.shared (LCOM 1).
- SplitField: methods split between self.x and self.y (LCOM 2).
- AbstractThing: abc.ABC subclass with one abstract method.
- Concrete: plain concrete class.
- _Internal: underscore-prefixed class.
- module_function: module-level function, excluded from class LCOM.
"""

import abc


class Cohesive:
    """Three methods all reading or writing self.shared (LCOM 1).

    The class deliberately has no __init__: every method reaches
    self.shared directly, so all three form a single connected
    component.
    """

    def set_shared(self, value: int) -> None:
        """Write self.shared."""
        self.shared = value

    def bump_shared(self) -> int:
        """Read and write self.shared."""
        self.shared = self.shared + 1
        return self.shared

    def read_shared(self) -> int:
        """Read self.shared."""
        return self.shared


class SplitField:
    """Four methods split between self.x and self.y (LCOM 2).

    set_x and read_x touch only self.x; set_y and read_y touch only
    self.y. No method touches both fields — and there is no __init__
    assigning both — so the methods form exactly two connected
    components. Adding a constructor that assigns self.x and self.y
    would merge the components and collapse the LCOM to 1; do not
    add one.
    """

    def set_x(self, value: int) -> None:
        """Write self.x only."""
        self.x = value

    def read_x(self) -> int:
        """Read self.x only."""
        return self.x

    def set_y(self, value: int) -> None:
        """Write self.y only."""
        self.y = value

    def read_y(self) -> int:
        """Read self.y only."""
        return self.y


class AbstractThing(abc.ABC):
    """Abstract base class carrying a single abstract method."""

    @abc.abstractmethod
    def perform(self) -> str:
        """Perform the thing."""


class Concrete:
    """Plain concrete class: no abstract marker, no underscore prefix."""

    def describe(self) -> str:
        """Return a stable description."""
        return "concrete"


class _Internal:
    """Underscore-prefixed class: private-name classification."""

    def hidden(self) -> str:
        """Return a stable hidden marker."""
        return "internal"


def module_function(value: int) -> int:
    """Module-level function; excluded from class LCOM analysis."""
    return value * 2

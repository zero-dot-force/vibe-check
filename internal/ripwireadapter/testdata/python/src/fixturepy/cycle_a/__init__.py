"""Cycle side A of the two-directory import cycle.

This package imports from fixturepy.cycle_b, which imports back from
fixturepy.cycle_a. Runtime importability is not required: the pair
exercises static import-graph cycle detection.
"""

from fixturepy.cycle_b.api import helper_b

__all__ = ["helper_b"]

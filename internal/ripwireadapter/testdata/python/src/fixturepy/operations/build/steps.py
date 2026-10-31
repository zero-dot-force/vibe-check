"""Build steps hosted inside the denylist-named 'build' directory."""

STEPS = ["fetch", "compile", "package"]


def run_steps() -> list:
    """Return a copy of the build steps defined by this module."""
    return list(STEPS)

"""Neo3-boa example contract for nsmith."""

from boa3.sc.compiletime import public


@public
def get_value() -> str:
    return "nsmith-python-ok"

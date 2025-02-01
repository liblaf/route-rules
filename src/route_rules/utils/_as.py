from collections.abc import Iterable


def as_set(obj: str | Iterable[str]) -> set[str]:
    if isinstance(obj, str):
        return {obj}
    return set(obj)

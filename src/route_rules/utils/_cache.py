from collections.abc import Awaitable, Callable
from typing import Any

from liblaf import grapes


@grapes.decorator
async def async_cached_method[T](
    wrapped: Callable[..., Awaitable[T]],
    instance: Any,
    args: tuple,
    kwargs: dict[str, Any],
) -> T:
    cache: dict[int, T] = getattr(instance, "_cache", {})
    key: int = hash((args, frozenset(kwargs.items())))
    if key in cache:
        return cache[key]
    result: T = await wrapped(*args, **kwargs)
    cache[key] = result
    object.__setattr__(instance, "_cache", cache)
    return result

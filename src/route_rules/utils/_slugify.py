from typing import Any

from slugify import slugify


def default_slug(self: Any) -> str:
    return slugify(self.name)

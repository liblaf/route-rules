import abc

import attrs

from route_rules.core import RuleSet


class Provider(abc.ABC):
    @abc.abstractmethod
    async def load(self) -> RuleSet: ...


@attrs.define
class ProviderFactory(abc.ABC):
    name: str = attrs.field()

    @abc.abstractmethod
    def create(self, name: str, /) -> Provider: ...

from typing import NamedTuple

from sbr import Rule, Source
from sbr.source.preset._const import PRESETS, get_preset


class PresetConfig(NamedTuple):
    id: str
    name: str
    include: list[str]
    exclude: list[str]


class Preset(Source):
    name: str = "preset"

    async def _get_nocache(self, key: str) -> Rule:
        return await get_preset(key)

    async def _keys_nocache(self) -> list[str]:
        return [preset.id for preset in PRESETS]

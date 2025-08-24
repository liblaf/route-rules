import datetime
import os
from pathlib import Path
from typing import Self

import anyio
import attrs
import prettytable

from route_rules.provider import Behavior, Format, ProviderFactoryRegistry

from ._schema import Config
from ._target import PrettyTarget

BEHAVIOR_FORMAT: list[tuple[Behavior, Format]] = [
    (Behavior.DOMAIN, Format.MRS),
    (Behavior.IPCIDR, Format.MRS),
    (Behavior.DOMAIN, Format.YAML),
    (Behavior.IPCIDR, Format.YAML),
    (Behavior.CLASSICAL, Format.YAML),
    (Behavior.DOMAIN, Format.TEXT),
    (Behavior.IPCIDR, Format.TEXT),
    (Behavior.CLASSICAL, Format.TEXT),
]


@attrs.define
class Builder:
    dist_dir: Path = attrs.field(default=Path("dist"))
    targets: list[PrettyTarget] = attrs.field(factory=list)

    @classmethod
    def load(cls, file: str | os.PathLike[str]) -> Self:
        config: Config = Config.load(file)
        factories: ProviderFactoryRegistry = ProviderFactoryRegistry.presets()
        return cls(
            targets=[
                PrettyTarget(
                    name=t.name, providers=t.providers, slug=slug, factories=factories
                )
                for slug, t in config.targets.items()
            ]
        )

    async def build(self) -> None:
        self.dist_dir.mkdir(parents=True, exist_ok=True)
        report: Path = self.dist_dir / "README.md"
        async with await anyio.open_file(report, "w") as fp:
            await fp.write("# Route Rules\n")
            await fp.write("<!-- body-start -->\n")
            now: datetime.datetime = datetime.datetime.now().astimezone()
            await fp.write(f"\nLast Updated At: {now.isoformat(timespec='seconds')}\n")
            for target in self.targets:
                await fp.write(f"\n## {target.name}\n")
                for behavior, format in BEHAVIOR_FORMAT:  # noqa: A001
                    await target.save(self.dist_dir, behavior=behavior, format=format)
                links: prettytable.PrettyTable = await target.pretty_links()
                await fp.write(f"\n{links.get_string()}\n")
                statistics: prettytable.PrettyTable = await target.pretty_statistics()
                await fp.write(f"\n{statistics.get_string()}\n")

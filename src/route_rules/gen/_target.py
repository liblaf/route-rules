from pathlib import Path

import attrs
import prettytable

from route_rules.core import RuleSet, Statistics, Target
from route_rules.provider import Behavior, Format, ProviderMihomo


@attrs.define
class Artifact:
    behavior: Behavior
    format: Format
    path: str
    size: int

    @property
    def name(self) -> str:
        return f"{self.format.upper()} {self.behavior.upper()}"


CDN: list[tuple[str, str]] = [
    ("GitHub", "https://raw.githubusercontent.com/liblaf/route-rules/dist/{path}"),
    ("jsDeliver", "https://cdn.jsdelivr.net/gh/liblaf/route-rules@dist/{path}"),
]


@attrs.define
class PrettyTarget(Target):
    artifacts: list[Artifact] = attrs.field(factory=list, kw_only=True)

    async def pretty_links(self) -> prettytable.PrettyTable:
        fields: list[str] = ["Format", "Behavior"]
        fields += (name for name, _ in CDN)
        table = prettytable.PrettyTable(fields, align="l")
        table.set_style(prettytable.TableStyle.MARKDOWN)
        for artifact in self.artifacts:
            row: list[str] = [artifact.format, artifact.behavior]
            for _, template in CDN:
                url: str = template.format(path=artifact.path)
                row.append(f"[Link]({url})")
            table.add_row(row)
        return table

    async def pretty_statistics(self) -> prettytable.PrettyTable:
        ruleset: RuleSet = await self.build()
        return Statistics.compare(await self.statistics(), ruleset.statistics)

    async def save(self, dist_dir: Path, behavior: Behavior, format: Format) -> None:  # noqa: A002
        ruleset: RuleSet = await self.build()
        path: str = f"mihomo/{format}/{behavior}/{self.slug}{format.ext}"
        size: int = ProviderMihomo.save(
            dist_dir / path, ruleset, behavior=behavior, format=format
        )
        if size <= 0:
            return
        self.artifacts.append(
            Artifact(behavior=behavior, format=format, path=path, size=size)
        )

    def _get_artifact_path(self, behavior: Behavior, format: Format) -> str:  # noqa: A002
        return f"mihomo/{format}/{behavior}/{self.slug}{format.ext}"

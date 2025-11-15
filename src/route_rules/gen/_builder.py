import collections
import os
from datetime import datetime
from pathlib import Path
from typing import Self

import attrs

from route_rules.core import RuleSet
from route_rules.export import ExporterMihomo
from route_rules.provider import Behavior, Format

from ._config import Config
from ._meta import ArtifactMeta, Meta, ProviderMeta, RecipeMeta, RecipeStatistics
from ._recipe import Recipe


def _default_exporters() -> list[ExporterMihomo]:
    return [
        ExporterMihomo(behavior=Behavior.DOMAIN, format=Format.MRS),
        ExporterMihomo(behavior=Behavior.DOMAIN, format=Format.TEXT),
        ExporterMihomo(behavior=Behavior.IPCIDR, format=Format.MRS),
        ExporterMihomo(behavior=Behavior.IPCIDR, format=Format.TEXT),
        ExporterMihomo(behavior=Behavior.CLASSICAL, format=Format.TEXT),
    ]


@attrs.define
class Builder:
    dist_dir: Path = Path("dist")
    exporters: list[ExporterMihomo] = attrs.field(factory=_default_exporters)
    recipes: list[Recipe] = attrs.field(factory=list)

    @classmethod
    def load(cls, file: str | os.PathLike[str]) -> Self:
        config: Config = Config.load(file)
        self: Self = cls()
        for recipe_config in config.recipes:
            self.recipes.append(Recipe.from_config(recipe_config))
        return self

    async def build(self) -> None:
        meta = Meta(build_time=datetime.now().astimezone())
        for recipe in self.recipes:
            meta.recipes.append(await self.build_recipe(recipe))
        meta_file: Path = self.dist_dir / "meta.json"
        meta_file.parent.mkdir(parents=True, exist_ok=True)
        meta_file.write_text(meta.json_encode())

    async def build_recipe(self, recipe: Recipe) -> RecipeMeta:
        ruleset: RuleSet = await recipe.build()
        meta = RecipeMeta(
            name=recipe.name,
            slug=recipe.slug,
            statistics=await self._build_statistics(recipe, ruleset),
        )
        for provider in recipe.providers:
            meta.providers.append(
                ProviderMeta(
                    name=provider,
                    download_url=recipe.registry.download_url(provider),
                    preview_url=recipe.registry.preview_url(provider),
                )
            )
        for exporter in self.exporters:
            file: Path | None = exporter.export(self.dist_dir, recipe.slug, ruleset)
            if file is None:
                continue
            meta.artifacts.append(
                ArtifactMeta(
                    behavior=exporter.behavior,
                    format=exporter.format,
                    path=file,
                    size=file.stat().st_size,
                )
            )
        return meta

    async def _build_statistics(
        self, recipe: Recipe, ruleset: RuleSet
    ) -> RecipeStatistics:
        outputs: dict[str, int] = {typ: len(values) for typ, values in ruleset.items()}
        inputs: dict[str, int] = collections.defaultdict(lambda: 0)
        for provider in recipe.providers:
            ruleset: RuleSet = await recipe.registry.load(provider)
            for typ, values in ruleset.items():
                inputs[typ] += len(values)
        return RecipeStatistics(inputs=inputs, outputs=outputs)

import urllib.parse
from pathlib import Path

import attrs
import prettytable

import route_rules as rr

CDN: list[tuple[str, str]] = [
    ("GitHub", "https://raw.githubusercontent.com/liblaf/route-rules/dist/{path}"),
    ("jsDeliver", "https://cdn.jsdelivr.net/gh/liblaf/route-rules@dist/{path}"),
]


@attrs.define
class Renderer:
    meta: rr.Meta = attrs.field()
    docs_dir: Path = attrs.field(default=Path("dist/docs"))
    site_docs_dir: Path = attrs.field(default=Path("docs/rulesets"))

    def render(self) -> None:
        for recipe in self.meta.recipes:
            self.render_recipe(recipe)

    def render_recipe(self, recipe: rr.RecipeMeta) -> None:
        self.docs_dir.mkdir(parents=True, exist_ok=True)
        filename: str = f"{recipe.slug}.md"
        with (self.docs_dir / filename).open("w") as fp:
            fp.write(f"# {recipe.name}\n")
            build_time: str = self.meta.build_time.isoformat(timespec="seconds")
            fp.write(f"Last Updated At: {build_time}\n")
            fp.write("## Links\n")
            links: prettytable.PrettyTable = self.render_links(recipe)
            fp.write(links.get_string() + "\n")
            fp.write("## Statistics\n")
            statistics: prettytable.PrettyTable = self.render_statistics(recipe)
            fp.write(statistics.get_string() + "\n")
            fp.write("## Sources\n")
            providers: prettytable.PrettyTable = self.render_providers(recipe)
            fp.write(providers.get_string() + "\n")
        self.site_docs_dir.mkdir(parents=True, exist_ok=True)
        url: str = (
            f"https://raw.githubusercontent.com/liblaf/route-rules/dist/docs/{filename}"
        )
        (self.site_docs_dir / filename).write_text(f'{{% include-markdown "{url}" %}}')

    def render_links(self, recipe: rr.RecipeMeta) -> prettytable.PrettyTable:
        field_names: list[str] = ["Name"]
        field_names += (name for name, _ in CDN)
        table = prettytable.PrettyTable(field_names, align="c")
        table.align["Name"] = "l"
        table.set_style(prettytable.TableStyle.MARKDOWN)
        for artifact in recipe.artifacts:
            row: list[str] = [artifact.path.name]
            for _, template in CDN:
                path: str = urllib.parse.quote(artifact.path.as_posix())
                url: str = template.format(path=path)
                row.append(f"[Link]({url})")
            table.add_row(row)
        return table

    def render_statistics(self, recipe: rr.RecipeMeta) -> prettytable.PrettyTable:
        table = prettytable.PrettyTable(["Type", "I", "O"], align="r")
        table.align["Type"] = "l"
        table.set_style(prettytable.TableStyle.MARKDOWN)
        inputs: dict[str, int] = recipe.statistics.inputs
        outputs: dict[str, int] = recipe.statistics.outputs
        for typ in sorted(inputs.keys() | outputs.keys()):
            table.add_row([typ, inputs.get(typ, 0), outputs.get(typ, 0)])
        inputs_total: int = sum(inputs.values())
        outputs_total: int = sum(outputs.values())
        table.add_row(["Total", inputs_total, outputs_total])
        return table

    def render_providers(self, recipe: rr.RecipeMeta) -> prettytable.PrettyTable:
        table = prettytable.PrettyTable(["Name", "Preview", "Download"], align="c")
        table.align["Name"] = "l"
        table.set_style(prettytable.TableStyle.MARKDOWN)
        for provider in recipe.providers:
            table.add_row(
                [
                    provider.name,
                    f"[Preview]({provider.preview_url})",
                    f"[Download]({provider.download_url})",
                ]
            )
        return table


def main() -> None:
    meta: rr.Meta = rr.Meta.json_decode(Path("dist/meta.json").read_bytes())
    renderer = Renderer(meta=meta)
    renderer.render()


if __name__ == "__main__":
    main()

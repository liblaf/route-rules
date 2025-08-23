import asyncio
import datetime
from pathlib import Path
from typing import TextIO

import prettytable
from liblaf import grapes

import route_rules as rr

DIST_DIR: Path = Path("dist")
HEADER: str = """\
# Route Rules

Last Updated At: {now}
"""


def save_target(target: rr.Target, rule_set: rr.RuleSet) -> None:
    rr.ProviderMihomo.save(
        DIST_DIR / f"mihomo/domain/yaml/{target.slug}.yaml",
        rule_set=rule_set,
        behavior=rr.Behavior.DOMAIN,
        format=rr.Format.YAML,
    )


async def safe_statistics(
    target: rr.Target, rule_set: rr.RuleSet, file: TextIO
) -> None:
    table: prettytable.PrettyTable = rr.Statistics.compare(
        await target.statistics(), rule_set.statistics
    )
    file.write(f"## {target.name}\n")
    file.write(table.get_string())
    file.write("\n")


async def main() -> None:
    grapes.logging.init()
    targets: list[rr.Target] = [
        rr.Target(
            name="🇨🇳 CN",
            providers=["MetaCubeX/geosite:cn", "MetaCubeX/geosite:geolocation-cn"],
        ),
        rr.Target(name="🎯 Local", providers=["MetaCubeX/geosite:private"]),
        rr.Target(name="🚀 Proxy", providers=["MetaCubeX/geosite:geolocation-!cn"]),
    ]
    DIST_DIR.mkdir(parents=True, exist_ok=True)
    with (DIST_DIR / "README.md").open("w") as fp:
        now: datetime.datetime = datetime.datetime.now().astimezone()
        fp.write(HEADER.format(now=now.isoformat(timespec="seconds")))
        for target in targets:
            rule_set: rr.RuleSet = await target.build()
            await safe_statistics(target, rule_set, file=fp)
            save_target(target, rule_set)


if __name__ == "__main__":
    asyncio.run(main())

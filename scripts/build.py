import asyncio
import datetime
from pathlib import Path
from typing import TextIO

import prettytable
from liblaf import grapes
from loguru import logger

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
    logger.disable("httpx")
    targets: list[rr.Target] = [
        rr.Target(
            name="🤖 AI",
            providers=[
                "dler-io:AI Suite",
                "MetaCubeX/geosite:category-ai-!cn",
                "SukkaW/classical:non_ip/ai",
            ],
        ),
        rr.Target(
            name="🎯 Direct",
            providers=[
                "blackmatrix7:Direct",
                "dler-io:Special",
                "SukkaW/classical:non_ip/direct",
            ],
        ),
        rr.Target(
            name="📍 Domestic",
            providers=[
                "blackmatrix7:ChinaMax",
                "dler-io:Domestic",
                "dler-io:Domestic IPs",
                "MetaCubeX/geosite:cn",
                "MetaCubeX/geosite:geolocation-cn",
                "SukkaW/classical:ip/domestic",
                "SukkaW/classical:non_ip/domestic",
            ],
        ),
        rr.Target(
            name="📥 Download",
            providers=[
                "SukkaW/classical:non_ip/cdn",
                "SukkaW/classical:non_ip/download",
                "SukkaW/domain:domainset/cdn",
                "SukkaW/domain:domainset/download",
            ],
        ),
        rr.Target(
            name="🌐 Global",
            providers=[
                "blackmatrix7:Global",
                "dler-io:Proxy",
                "MetaCubeX/geosite:geolocation-!cn",
                "SukkaW/classical:non_ip/global",
            ],
        ),
        rr.Target(
            name="♾️ Local",
            providers=[
                "blackmatrix7:Lan",
                "dler-io:LAN",
                "MetaCubeX/geosite:private",
                "SukkaW/classical:ip/lan",
                "SukkaW/classical:non_ip/lan",
            ],
        ),
        rr.Target(
            name="📺 Stream",
            providers=[
                "blackmatrix7:GlobalMedia",
                "SukkaW/classical:ip/stream",
                "SukkaW/classical:non_ip/stream",
            ],
        ),
    ]
    DIST_DIR.mkdir(parents=True, exist_ok=True)
    with (DIST_DIR / "statistics.md").open("w") as fp:
        now: datetime.datetime = datetime.datetime.now().astimezone()
        fp.write(HEADER.format(now=now.isoformat(timespec="seconds")))
        for target in targets:
            rule_set: rr.RuleSet = await target.build()
            await safe_statistics(target, rule_set, file=fp)
            save_target(target, rule_set)


if __name__ == "__main__":
    asyncio.run(main())

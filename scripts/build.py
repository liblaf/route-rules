import asyncio
import dataclasses
import datetime
import functools
import pathlib
from collections.abc import Callable, Coroutine
from typing import TextIO

from icecream import ic
from sbr import Rule, preset


@dataclasses.dataclass(kw_only=True)
class Config:
    fn: Callable[[], Coroutine[None, None, Rule]]
    geosite: bool = False
    id: str
    name: str


RULE_SETS: list[Config] = [
    Config(id="ads", fn=preset.ads, name="📵 RuleSet:ADs", geosite=True),
    Config(id="private", fn=preset.private, name="🔒 RuleSet:Private", geosite=True),
    Config(id="cn", fn=preset.cn, name="🇨🇳 RuleSet:CN", geosite=True),
    Config(id="ai", fn=preset.ai, name="🤖 RuleSet:AI"),
    Config(id="emby", fn=preset.emby, name="🍟 RuleSet:Emby"),
    Config(id="download", fn=preset.download, name="☁️ RuleSet:Download"),
    Config(id="media", fn=preset.media, name="📺 RuleSet:Media"),
]


async def main() -> None:
    summary_filename = pathlib.Path("output/README.md")
    summary_filename.parent.mkdir(parents=True, exist_ok=True)
    with summary_filename.open("w") as fp:  # noqa: ASYNC230
        fprint = functools.partial(print, file=fp)
        fprint("# sing-box Rules")
        now: datetime.datetime = datetime.datetime.now(datetime.UTC)
        fprint("Updated At:", now.strftime("%Y-%m-%d %H:%M:%S"))
        for cfg in RULE_SETS:
            raw: Rule = await cfg.fn()
            rule: Rule = raw.model_copy(deep=True)
            rule.optimize()
            rule.save(f"output/rule-set/{cfg.id}.json")
            raw.save(f"output/rule-set/{cfg.id}-raw.json")
            print_summary(cfg.name, raw, rule, fp)
            ic(cfg.name, rule)
            if cfg.geosite:
                rule.ip_cidr.clear()
                rule.save(f"output/geosite/{cfg.id}.json")


def print_summary(name: str, raw: Rule, rule: Rule, file: TextIO) -> None:
    fprint = functools.partial(print, file=file)
    fprint("##", name)
    fprint("| Type | Count (Raw) | Count (Opt) |")
    fprint("| ---- | ----------: | ----------: |")
    if raw.domain:
        fprint("| DOMAIN |", len(raw.domain), "|", len(rule.domain), "|")
    if raw.domain_suffix:
        fprint(
            "| DOMAIN-SUFFIX |",
            len(raw.domain_suffix),
            "|",
            len(rule.domain_suffix),
            "|",
        )
    if raw.domain_keyword:
        fprint(
            "| DOMAIN-KEYWORD |",
            len(raw.domain_keyword),
            "|",
            len(rule.domain_keyword),
            "|",
        )
    if raw.domain_regex:
        fprint(
            "| DOMAIN-REGEX |", len(raw.domain_regex), "|", len(rule.domain_regex), "|"
        )
    if raw.ip_cidr:
        fprint("| IP-CIDR |", len(raw.ip_cidr), "|", len(rule.ip_cidr), "|")
    if raw.process_name:
        fprint(
            "| PROCESS-NAME |", len(raw.process_name), "|", len(rule.process_name), "|"
        )
    fprint("| TOTAL |", len(raw), "|", len(rule), "|")


if __name__ == "__main__":
    asyncio.run(main())

import asyncio
import datetime

from sbr import Rule
from sbr.typing import StrPath

RULE_SETS: dict[str, str] = {
    "📵 RuleSet:ADs": "output/rule-set/ads.json",
    "🔒 RuleSet:Private": "output/rule-set/private.json",
    "🇨🇳 RuleSet:CN": "output/rule-set/cn.json",
    "🤖 RuleSet:AI": "output/rule-set/ai.json",
    "📺 RuleSet:Media": "output/rule-set/media.json",
    "☁️ RuleSet:Download": "output/rule-set/download.json",
    "🍟 RuleSet:Emby": "output/rule-set/emby.json",
}


async def summary(name: str, filename: StrPath) -> str:
    result: str = f"## {name}\n"
    rule: Rule = await Rule.from_json_url(filename)
    result += "| Type | Count |\n"
    result += "| ---- | ----: |\n"
    if rule.domain:
        result += f"| DOMAIN | {len(rule.domain)} |\n"
    if rule.domain_suffix:
        result += f"| DOMAIN-SUFFIX | {len(rule.domain_suffix)} |\n"
    if rule.domain_keyword:
        result += f"| DOMAIN-KEYWORD | {len(rule.domain_keyword)} |\n"
    if rule.domain_regex:
        result += f"| DOMAIN-REGEX | {len(rule.domain_regex)} |\n"
    if rule.ip_cidr:
        result += f"| IP-CIDR | {len(rule.ip_cidr)} |\n"
    if rule.process_name:
        result += f"| PROCESS-NAME | {len(rule.process_name)} |\n"
    result += f"| TOTAL | {len(rule)} |\n"
    return result


async def main() -> None:
    print("# sing-box Rules")
    print(
        "Updated At:", datetime.datetime.now(datetime.UTC).strftime("%Y-%m-%d %H:%M:%S")
    )
    for name, filename in RULE_SETS.items():
        print(await summary(name, filename))


if __name__ == "__main__":
    asyncio.run(main())

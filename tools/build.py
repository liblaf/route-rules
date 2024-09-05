import asyncio
import datetime
from pathlib import Path

import anyio
import prettytable
from prettytable import PrettyTable

import sbr
from sbr import PRESETS, Rule, Source


async def gen_optimization_summary(preset: Source) -> dict[str, Rule]:
    fpath: Path = Path("output/README.md")
    fpath.parent.mkdir(parents=True, exist_ok=True)
    rules: dict[str, Rule] = {}
    async with await anyio.open_file(fpath, "w") as fp:
        await fp.write("# sing-box Rules\n")
        now: datetime.datetime = datetime.datetime.now(datetime.UTC)
        await fp.write(f"Updated at: {now.isoformat()}\n")
        for cfg in PRESETS:
            rule_raw: Rule = await preset.get(cfg.id)
            rule_opt: Rule = rule_raw.model_copy(deep=True)
            rule_opt.optimize()
            rules[cfg.id] = rule_opt
            table: PrettyTable = PrettyTable(["Type", "Count (Raw)", "Count (Opt)"])
            table.align.update({"Type": "l", "Count (Raw)": "r", "Count (Opt)": "r"})
            for k, v in rule_raw:
                name: str = k.upper().replace("_", "-")
                table.add_row([name, len(v), len(rule_opt[k])])
            table.add_row(["TOTAL", len(rule_raw), len(rule_opt)])
            await fp.write(f"## {cfg.name}\n")
            table.set_style(prettytable.MARKDOWN)
            await fp.write(table.get_string())
            await fp.write("\n")
    return rules


async def main() -> None:
    preset: Source = sbr.get_source("preset")
    rules: dict[str, Rule] = await gen_optimization_summary(preset)
    for k, r in rules.items():
        r.save(f"output/rule-set/{k}.json")
        r.geoip().save(f"output/geoip/{k}.json")
        r.geosite().save(f"output/geosite/{k}.json")


if __name__ == "__main__":
    sbr.logging.init()
    asyncio.run(main())

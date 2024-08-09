import asyncio

from icecream import ic
from sbr import Rule


async def main() -> None:
    rule = Rule()
    rule += await Rule.from_json_url("data/NotSFC/Emby.json")
    ic(rule)
    rule.save("output/rule-set/emby.json")


if __name__ == "__main__":
    asyncio.run(main())

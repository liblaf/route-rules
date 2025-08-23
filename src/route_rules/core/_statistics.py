import collections
from collections.abc import Mapping
from typing import Self

import prettytable


class Statistics(collections.UserDict[str, int]):
    def __add__(self, other: Mapping[str, int], /) -> Self:
        result: Self = type(self)()
        for key in self.keys() | other.keys():
            result[key] = self.get(key, 0) + other.get(key, 0)
        return result

    @classmethod
    def compare(
        cls, before: Mapping[str, int], after: Mapping[str, int]
    ) -> prettytable.PrettyTable:
        table = prettytable.PrettyTable(["Type", "w/o Optim", "w/ Optim"])
        table.align.update({"Type": "l", "w/o Optim": "r", "w/ Optim": "r"})
        table.set_style(prettytable.TableStyle.MARKDOWN)
        for key in sorted(before.keys() | after.keys()):
            before_len: int = before.get(key, 0)
            after_len: int = after.get(key, 0)
            if not (before_len and after_len):
                continue
            name: str = key.replace("_", "-").upper()
            table.add_row([name, before_len, after_len])
        before_total: int = sum(before.values())
        after_total: int = sum(after.values())
        table.add_row(["TOTAL", before_total, after_total])
        return table
